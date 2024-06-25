#!/bin/bash

#####################################################################################
# Summary:
# This script is designed to create and clean up a Kind (Kubernetes IN Docker) cluster
# named 'k8s-integration-tests'.
#
# Requirements:
# - kubectl version: Client Version: v1.30.2, Kustomize Version: v5.0.4 or compatible
# - kind version: kind v0.23.0 or compatible
# - docker version: 27.0.1
#####################################################################################

#####################################################################################
# Function to check if a command is available
check_command_available() {
  command -v $1 >/dev/null 2>&1 || { echo >&2 "$1 is required but it's not installed. Aborting."; exit 1; }
}

#####################################################################################
# Function to check the result of the last command and exit if it failed
check_command() {
  if [ $? -ne 0 ]; then
    echo "$1"
    cleanup_kind_cluster
    exit 1
  fi
}

#####################################################################################
# Function to create Kind cluster
create_kind_cluster() {
  echo "Creating Kind cluster..."
  kind create cluster --name k8s-integration-tests
  check_command "Failed to create Kind cluster."
  echo "Kind cluster created successfully."
}

#####################################################################################
# Function to cleanup Kind cluster
cleanup_kind_cluster() {
  echo "Deleting Kind cluster..."

  # Ensure the port-forward process is killed
  if [ -n "$PORT_FORWARD_PID" ]; then
    echo "Killing port-forward process with PID $PORT_FORWARD_PID..."
    kill $PORT_FORWARD_PID
  fi

  kind delete cluster --name k8s-integration-tests

  # Check if the cluster was deleted successfully
  if [ $? -ne 0 ]; then
    echo "Failed to delete Kind cluster. Attempting to remove container manually..."

    # Try to remove the control-plane container manually
    docker rm -f -v k8s-integration-tests-control-plane
    if [ $? -ne 0 ]; then
      echo "Manual removal of the control-plane container failed. Please check the container state and logs."
      exit 1
    fi

    echo "Control-plane container removed manually."
  else
    echo "Kind cluster deleted successfully."
  fi
}

#####################################################################################
# Function to build Docker images locally
build_docker_images() {
    echo "Building receiver Docker image..."
    docker build -t receiver:unique-tag -f ../test/integration/docker/receiver/Dockerfile ../.
    check_command "Failed to build receiver Docker image."

    echo "Building transmitter Docker image..."
    docker build -t transmitter:unique-tag -f ../test/integration/docker/transmitter/Dockerfile ../.
    check_command "Failed to build transmitter Docker image."

    echo "Building Tor Docker image..."
    docker build -t tor-proxy:unique-tag -f ../test/integration/docker/tor/Dockerfile ../test/integration/docker/tor
    check_command "Failed to build tor proxy Docker image."

    echo "Docker images built successfully."
}

#####################################################################################
# Function to check if Kind cluster is running
check_kind_cluster_running() {
  echo "Checking if Kind cluster is running..."
  clusters=$(kind get clusters)
  if [[ "$clusters" == "" ]]; then
    echo "No Kind clusters are currently running."
    exit 1
  else
    echo "The following Kind clusters are running:"
    echo "$clusters"
  fi

  if [[ "$clusters" != *"k8s-integration-tests"* ]]; then
    echo "The 'k8s-integration-tests' cluster is not running."
    exit 1
  fi

  echo "Kind cluster 'k8s-integration-tests' is running."
}

#####################################################################################
# Function to load Docker images into Kind cluster
load_docker_images() {
  echo "Loading receiver Docker image into Kind cluster..."
  kind load docker-image receiver:unique-tag --name k8s-integration-tests
  check_command "Failed to load receiver Docker image into Kind cluster."

  echo "Loading transmitter Docker image into Kind cluster..."
  kind load docker-image transmitter:unique-tag --name k8s-integration-tests
  check_command "Failed to load transmitter Docker image into Kind cluster."

  echo "Loading tor proxy Docker image into Kind cluster..."
  kind load docker-image tor-proxy:unique-tag --name k8s-integration-tests
  check_command "Failed to load tor proxy Docker image into Kind cluster."

  echo "Docker images loaded into Kind cluster successfully."
}

#####################################################################################
# Function to deploy Kubernetes resources
deploy_kubernetes_resources() {
  echo "Deploying Kubernetes resources..."
  kubectl apply -f ../.github/workflows/k8s/full/receiver-deployment.yml
  check_command "Failed to deploy receiver deployment."

  kubectl apply -f ../.github/workflows/k8s/rabbitmq-deployment.yml
  check_command "Failed to deploy RabbitMQ deployment."

  kubectl apply -f ../.github/workflows/k8s/tor-deployment.yml
  check_command "Failed to deploy Tor deployment."
}

#####################################################################################
# Function to validate deployments
validate_deployments() {
    sleep 60
    echo "Waiting for receiver pods to be ready..."
    kubectl wait --namespace default --for=condition=ready pod --selector=app=receiver --timeout=900s
    check_command "Receiver pods did not become ready in time."

    echo "Waiting for RabbitMQ pods to be ready..."
    kubectl wait --namespace default --for=condition=ready pod -l app=rabbitmq --timeout=240s
    check_command "RabbitMQ pods did not become ready in time."

    echo "Waiting for Tor pods to be ready..."
    kubectl wait --namespace default --for=condition=ready pod -l app=tor --timeout=120s
    check_command "Tor pods did not become ready in time."

    echo "All deployments are ready."
}

#####################################################################################
# Function to expose receiver service with ngrok using Docker
expose_receiver_with_ngrok() {
    echo "Stopping any existing ngrok container..."
    docker stop ngrok || true

    echo "Starting ngrok using Docker..."
    docker run -d --name ngrok --rm --net=host -e NGROK_AUTHTOKEN=$NGROK_AUTHTOKEN ngrok/ngrok http 8080 --log=stdout
    check_command "Failed to start ngrok container."

    echo "Waiting for ngrok to start..."
    sleep 5 # Initial wait time for ngrok to start

    MAX_RETRIES=10
    RETRY_COUNT=0
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
        echo "Fetching ngrok URL (Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES)..."
        NGROK_URL=$(curl -s http://localhost:4040/api/tunnels | grep -oP 'https://[0-9a-zA-Z-]*\.ngrok[^"]*' | head -n 1)
        if [[ "$NGROK_URL" == https* ]]; then
            break
        fi
        RETRY_COUNT=$((RETRY_COUNT + 1))
        sleep 2
    done

    echo "Forwarding receiver service to local port 8080"
    kubectl port-forward svc/receiver-service 8080:80 &
    PORT_FORWARD_PID=$!
    sleep 5 # Give port-forwarding time to establish

    if [ -z "$NGROK_URL" ]; then
        echo "Failed to retrieve ngrok URL."
        kill $PORT_FORWARD_PID
        exit 1
    fi

    echo "Test URL: $NGROK_URL"
    export RECEIVER_URL=$NGROK_URL

    # Ensure port forwarding process is still running
    if ! kill -0 $PORT_FORWARD_PID > /dev/null 2>&1; then
        echo "Port forwarding process has terminated unexpectedly."
        exit 1
    fi

    echo "Port forwarding established successfully."
}

#####################################################################################
# Function to deploy transmitter
deploy_transmitter() {
    expose_receiver_with_ngrok

    TEMP_FILE=$(mktemp)
    envsubst < ../.github/workflows/k8s/full/transmitter-deployment.yml > "$TEMP_FILE"
    kubectl apply -f "$TEMP_FILE"
    check_command "Failed to deploy transmitter deployment."

    # Remove the temporary file
    rm -f "$TEMP_FILE"
}

#####################################################################################
# Function to handle script exit
handle_exit() {
  echo "Exiting script..."
  cleanup_kind_cluster
  echo "Cleanup completed. Exiting."
}


#####################################################################################
# Main script execution

export NGROK_AUTHTOKEN=12345

# Check for required commands
check_command_available kubectl
check_command_available kind
check_command_available docker

# Ensure NGROK_AUTHTOKEN is set
if [ -z "$NGROK_AUTHTOKEN" ]; then
  echo "NGROK_AUTHTOKEN environment variable is not set. Aborting."
  exit 1
fi

# Trap to handle script exit
trap handle_exit EXIT

create_kind_cluster
build_docker_images
check_kind_cluster_running
load_docker_images
deploy_kubernetes_resources
validate_deployments
deploy_transmitter

echo "Script completed successfully."

# Wait for Ctrl+C (SIGINT) to exit
echo "Press Ctrl+C to clean up and exit."
while :; do
  sleep 1
done
