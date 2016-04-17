#!/bin/bash

# Travis CI doesn't give us tun tap support and nothing else gives us root.
# These tests should only be run locally.

sudo -E env "PATH=$PATH" go test -v github.com/AutoRoute/node/integration_tests/root -race
