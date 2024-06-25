# Start from the official Golang base image
FROM golang:1.22 AS builder

WORKDIR /app

# Copy the go.mod and go.sum files from the root of the project
COPY go.mod go.sum ./

# Download all the dependencies.
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o receiver ./cmd/receiver/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/receiver .

# Expose the port the application will run on
EXPOSE 8888

# Run the application
CMD ["./receiver"]