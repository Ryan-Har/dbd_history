# Stage 1: Build the Go application
FROM golang:tip-alpine AS builder

# Set working directory for the build
WORKDIR /app
RUN apk add --no-cache \
    build-base \
    sqlite-dev

# Copy the Go module files and download dependencies
# This caches dependencies if go.mod/go.sum don't change
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/main .

# Stage 2: Create the final lean image
FROM alpine:latest

# Set working directory for the final application
WORKDIR /app

# Make directory for the DB
RUN mkdir db

# Copy the built executable from the builder stage
COPY --from=builder /app/main .

# Copy the static assets and templates
# Adjust these paths if your static/templates folders are not directly in the root of your Go app
COPY static/ ./static/
COPY templates/ ./templates/

# Expose the port your Go application listens on
EXPOSE 8080

# Command to run the application
# Use the absolute path to your executable
CMD ["/app/main"]