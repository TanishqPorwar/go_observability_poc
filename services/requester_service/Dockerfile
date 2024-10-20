FROM golang:1.23-bookworm as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
# Copy local code to the container image.
COPY ../../services/requester_service ./

#RUN #go mod tidy
# Build the binary.
RUN go build -o requester ./requester.go

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/requester /usr/local/bin/requester

# Expose port 8080 to the outside world
EXPOSE 8081
EXPOSE 6061

# Command to run the executable
CMD ["requester"]