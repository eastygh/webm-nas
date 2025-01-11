#!/bin/sh

docker run --rm  --platform=linux/arm/v7 -v $(pwd):/opt golang:1.23-bookworm sh -c "cd /opt && CGO_ENABLED=1 go build -ldflags \"-s -w\" -o server"
