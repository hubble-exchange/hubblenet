
#!/usr/bin/env bash
set -e
source ./scripts/utils.sh

if ! [[ "$0" =~ scripts/run_local.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

export VALIDATOR_PRIVATE_KEY="31b571bf6894a248831ff937bb49f7754509fe93bbd2517c9c73c4144c0e97dc"
if [[ -z "${VALIDATOR_PRIVATE_KEY}" ]]; then
  echo "VALIDATOR_PRIVATE_KEY must be set"
  exit 255
fi

avalanche network clean

./scripts/build.sh custom_evm.bin

avalanche subnet create hubblenet --force --custom --genesis genesis.json --vm custom_evm.bin --config .avalanche-cli.json

# configure and add chain.json
avalanche subnet configure hubblenet --chain-config chain.json --config .avalanche-cli.json
# avalanche subnet configure hubblenet --per-node-chain-config node_config.json --config .avalanche-cli.json

# use the same avalanchego version as the one used in subnet-evm
# use tee to keep showing outut while storing in a var
avalanche subnet deploy hubblenet -l --avalanchego-version v1.9.14 --config .avalanche-cli.json

setStatus
