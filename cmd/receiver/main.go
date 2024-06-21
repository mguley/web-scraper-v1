package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// handleRequests processes each HTTP request by logging the User-Agent and IP address, and responding with this
// information. This function acts as an HTTP handler.
//
// Parameters:
// - writer http.ResponseWriter: The writer to send data to the client.
// - request *http.Request: The HTTP request object containing all details of the request.
func handleRequests(writer http.ResponseWriter, request *http.Request) {
	userAgent := request.Header.Get("User-Agent")
	ipAddress := request.RemoteAddr

	log.Printf("Received request from IP: %s, User-Agent: %s", ipAddress, userAgent)

	// Construct and send the response containing the User-Agent and IP address
	responseMessage := "Received User-Agent: " + userAgent + "\nIP Address: " + ipAddress
	if _, err := writer.Write([]byte(responseMessage)); err != nil {
		log.Printf("Error writing response: %s", err)
		http.Error(writer, "Failed to write response", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

// startServer initializes and starts an HTTP server listening on the specified address.
// It registers handleRequests as the handler for the root path.
//
// Parameters:
// - address string: The network address that the server should listen on.
//
// Returns:
// - *http.Server: A pointer to the initialized server.
func startServer(address string) *http.Server {
	http.HandleFunc("/", handleRequests)

	server := &http.Server{Addr: address}

	go func() {
		log.Println("Starting server on", address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not start server: %v", err)
		}
	}()

	return server
}

// shutdownServer gracefully shuts down the provided HTTP server within the specified timeout.
// It listens for system interrupt or termination signals to trigger a shutdown.
//
// Parameters:
// - server *http.Server: The server to shut down.
// - timeout time.Duration: The maximum duration to wait for a graceful shutdown.
func shutdownServer(server *http.Server, timeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Block until a shutdown signal is received
	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func main() {
	server := startServer(":8888")
	shutdownServer(server, 10*time.Second)
}
