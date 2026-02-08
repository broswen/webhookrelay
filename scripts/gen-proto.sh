#!/bin/bash

# Script to generate Go code from proto files

set -e

# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Install required tools if not already installed
echo "Installing protoc-gen-go..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

echo "Installing protoc-gen-go-grpc..."
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from proto files
echo "Generating Go code from proto files..."
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  pkg/api/v1/*.proto

echo "Proto generation complete!"
