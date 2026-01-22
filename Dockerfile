# --- Stage 1: The Builder ---
# We use the official Go image to compile your code
FROM golang:1.25.6-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the app! 
# We disable CGO for a pure static binary that runs anywhere
RUN CGO_ENABLED=0 GOOS=linux go build -o portfolio main.go

# --- Stage 2: The Runner ---
# We use a tiny "Alpine" Linux image for the final app
FROM alpine:latest

# Install SSH client (sometimes needed for keys) and ca-certificates
RUN apk add --no-cache openssh-client ca-certificates

WORKDIR /root/

# Copy the binary from the Builder stage
COPY --from=builder /app/portfolio .

# Create a folder for the SSH keys to live in
RUN mkdir .ssh

# --- THE FIX IS HERE ---
# Force the terminal to advertise 256-color and TrueColor support
ENV TERM=xterm-256color
ENV COLORTERM=truecolor

# Open the port we chose in main.go
EXPOSE 23234

# Run the app
CMD ["./portfolio"]