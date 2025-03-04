#!/bin/bash

# Build for Linux
echo "Building for Linux/amd64..."
GOOS=linux GOARCH=amd64 go build -o pocketbase
if [ $? -eq 0 ]; then
    echo "✓ Linux build successful"
else
    echo "× Linux build failed"
fi

chmod +x pocketbase