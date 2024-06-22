# Start from the official Golang base image
FROM golang:1.22 as builder

WORKDIR /app
# Copy the go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Copy the source code into the container
#COPY cmd/transmitter .
# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/transmitter

#FROM alpine:latest
#WORKDIR /root/
#COPY --from=builder /app/main .
# Run the application
#CMD ["./main"]
