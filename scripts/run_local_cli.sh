
#!/usr/bin/env bash
set -e

if ! [[ "$0" =~ scripts/run_local_cli.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

if [[ -z "${ARTIFACT_PATH_PREFIX}" ]]; then
  echo "ARTIFACT_PATH_PREFIX must be set"
  exit 255
fi

avalanche network clean

./scripts/build.sh custom_evm.bin

avalanche subnet create hubblenet --force --custom --genesis genesis.json --vm custom_evm.bin

# configure and add chain.json
avalanche subnet configure hubblenet --chain-config chain.json

# use the same avalanchego version as the one used in subnet-evm
avalanche subnet deploy hubblenet -l --avalanchego-version v1.9.7
