#!/usr/bin/env bash

set -eu

go build -ldflags '-s -w'
mv update-shell-utils $GOPATH/bin
