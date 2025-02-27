#!/bin/bash

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o pocketbase
if [ $? -eq 0 ]; then
    chmod +x builds/elimt-cli-linux
    echo "✓ Linux build successful"
else
    echo "× Linux build failed"
fi

chmod +x pocketbase