#!/bin/bash

# This script compiles the OrbitDeploy application for Linux (amd64).

# --- Configuration ---
BINARY_NAME="orbit-deploy"

# --- Frontend Build ---
echo "Building frontend..."
cd frontend
if ! command -v bun &> /dev/null; then
  echo "bun is not installed. Please install it first." >&2
  exit 1
fi
bun install
bun run build
cd ..

# --- Backend Build ---
echo "Building OrbitDeploy for Linux (amd64)..."

# Set the target OS and architecture
export GOOS=linux
export GOARCH=amd64

# Build the application
go build -o "${BINARY_NAME}" main.go

# --- Post-build ---
if [ $? -eq 0 ]; then
  echo ""
  echo "Build successful!"
  echo "The binary '${BINARY_NAME}' has been created in the current directory."
  echo "You can now run the install.sh script to install it as a service."
else
  echo ""
  echo "Build failed. Please check the error messages above." >&2
  exit 1
fi
