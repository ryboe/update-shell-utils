#!/usr/bin/env bash

set -euxo pipefail

CGO_ENABLED=0 go install -ldflags '-s -w'
