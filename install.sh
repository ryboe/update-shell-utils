#!/usr/bin/env bash

set -euxo pipefail

CGO_ENABLED=0 go install -mod vendor -ldflags '-s -w'
