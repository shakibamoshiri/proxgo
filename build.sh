#!/bin/bash

set -eux

GOARCH=$(go env GOARCH) GOOS=$(go env GOOS) CGO_ENABLED=0 go build -ldflags="-s -w" -o prox .
