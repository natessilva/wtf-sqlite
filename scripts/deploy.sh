#!/bin/sh -e

REMOTE_HOST="silva.world"
SERVICE_NAME="dial"

echo "Building the Go binary..."
CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -o server -ldflags "-linkmode external -extldflags -static" cmd/server/main.go

echo "Copying the binary to the server..."
scp -q server "$REMOTE_HOST:~/server_new"

echo "Restarting the service on the server..."
ssh "$REMOTE_HOST" 'bash -s' << EOF
    set -e

    echo "Stopping the service..."
    sudo systemctl stop "$SERVICE_NAME"

    echo "Replacing the binary..."
    sudo cp -f server_new server

    echo "Starting the service..."
    sudo systemctl start "$SERVICE_NAME"
EOF

rm server
echo "Deployment completed successfully."
