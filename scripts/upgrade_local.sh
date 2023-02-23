#!/usr/bin/env bash
set -e

avalanche network stop --snapshot-name snap1

./scripts/build.sh custom_evm.bin

avalanche subnet upgrade vm hubblenet --binary custom_evm.bin --local

avalanche network start --avalanchego-version v1.9.7 --snapshot-name snap1
