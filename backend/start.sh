#!/bin/bash
export PATH="/usr/local/go/bin:$PATH"
export GOROOT="/usr/local/go"
export GOPATH="$HOME/go"

cd "$(dirname "$0")"

echo "Building shelfarr..."
go build ./cmd/shelfarr || exit 1

echo "Starting shelfarr with AUTH_DISABLED=true..."
AUTH_DISABLED=true ./shelfarr
