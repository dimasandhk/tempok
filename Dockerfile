FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o tempok .

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Add ca-certificates in case we need to make HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/tempok .

# Expose control and public ports (can be overridden, but these are typical defaults)
EXPOSE 9999
EXPOSE 80

# Run the executable
ENTRYPOINT ["./tempok"]
