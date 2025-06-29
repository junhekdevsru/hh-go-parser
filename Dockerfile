# Multi-stage build for hh.ru Resume Parser
FROM golang:1.21-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod file
COPY go.mod ./

# Download dependencies (if any)
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hh-parser main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh parser

# Set working directory
WORKDIR /home/parser

# Copy binary from builder stage
COPY --from=builder /app/hh-parser .

# Change ownership
RUN chown parser:parser hh-parser
RUN chmod +x hh-parser

# Switch to non-root user
USER parser

# Set default command
ENTRYPOINT ["./hh-parser"]

# Default help command
CMD ["-h"]
