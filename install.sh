#!/usr/bin/env bash

set -euxo pipefail

GO111MODULE=on CGO_ENABLED=0 go install -mod vendor -ldflags '-s -w'
