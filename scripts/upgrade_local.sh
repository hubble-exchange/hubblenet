#!/usr/bin/env bash
set -e
source ./scripts/utils.sh

./scripts/build.sh custom_evm.bin

avalanche network stop --snapshot-name snap1

avalanche subnet upgrade vm localnet --binary custom_evm.bin --local

# utse tee to keep showing outut while storing in a var
OUTPUT=$(avalanche network start --avalanchego-version v1.11.2 --snapshot-name snap1 --config .avalanche-cli.json | tee /dev/fd/2)

setStatus
