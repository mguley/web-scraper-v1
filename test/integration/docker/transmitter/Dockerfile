# Start from the official Golang base image
FROM golang:1.22 as builder

WORKDIR /app

# Copy the go.mod and go.sum files from the root of the project
COPY go.mod go.sum ./

# Download all the dependencies.
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transmitter ./cmd/transmitter/main.go

# Use a minimal base image to run the application
FROM alpine:latest

# Set the working directory in the new image
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/transmitter .

# Run the binary
CMD ["./transmitter"]
