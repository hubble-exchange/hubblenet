#!/usr/bin/env bash
set -e

if ! [[ "$0" =~ scripts/run_local.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

VERSION='v1.9.2'
ANR_VERSION='35be10cd3867a94fbe960a1c14a455f179de60d9'

# Load the versions
SUBNET_EVM_PATH=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  cd .. && pwd
)

# Load the constants
source "$SUBNET_EVM_PATH"/scripts/constants.sh


############################
# download avalanchego
# https://github.com/ava-labs/avalanchego/releases
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
BASEDIR=/tmp/subnet-evm-runner
mkdir -p ${BASEDIR}
AVAGO_DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/${VERSION}/avalanchego-linux-${GOARCH}-${VERSION}.tar.gz
AVAGO_DOWNLOAD_PATH=${BASEDIR}/avalanchego-linux-${GOARCH}-${VERSION}.tar.gz
if [[ ${GOOS} == "darwin" ]]; then
  AVAGO_DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/${VERSION}/avalanchego-macos-${VERSION}.zip
  AVAGO_DOWNLOAD_PATH=${BASEDIR}/avalanchego-macos-${VERSION}.zip
fi

AVAGO_FILEPATH=${BASEDIR}/avalanchego-${VERSION}
if [[ ! -d ${AVAGO_FILEPATH} ]]; then
  if [[ ! -f ${AVAGO_DOWNLOAD_PATH} ]]; then
    echo "downloading avalanchego ${VERSION} at ${AVAGO_DOWNLOAD_URL} to ${AVAGO_DOWNLOAD_PATH}"
    curl -L ${AVAGO_DOWNLOAD_URL} -o ${AVAGO_DOWNLOAD_PATH}
  fi
  echo "extracting downloaded avalanchego to ${AVAGO_FILEPATH}"
  if [[ ${GOOS} == "linux" ]]; then
    mkdir -p ${AVAGO_FILEPATH} && tar xzvf ${AVAGO_DOWNLOAD_PATH} --directory ${AVAGO_FILEPATH} --strip-components 1
  elif [[ ${GOOS} == "darwin" ]]; then
    unzip ${AVAGO_DOWNLOAD_PATH} -d ${AVAGO_FILEPATH}
    mv ${AVAGO_FILEPATH}/build/* ${AVAGO_FILEPATH}
    rm -rf ${AVAGO_FILEPATH}/build/
  fi
  find ${BASEDIR}/avalanchego-${VERSION}
fi

AVALANCHEGO_PATH=${AVAGO_FILEPATH}/avalanchego
AVALANCHEGO_PLUGIN_DIR=${AVAGO_FILEPATH}/plugins


#################################
# compile subnet-evm
# Check if SUBNET_EVM_COMMIT is set, if not retrieve the last commit from the repo.
# This is used in the Dockerfile to allow a commit hash to be passed in without
# including the .git/ directory within the Docker image.
subnet_evm_commit=${SUBNET_EVM_COMMIT:-$(git rev-list -1 HEAD)}

# Build Subnet EVM, which is run as a subprocess
echo "Building Subnet EVM Version: $subnet_evm_version; GitCommit: $subnet_evm_commit"
go build \
  -ldflags "-X github.com/ava-labs/subnet_evm/plugin/evm.GitCommit=$subnet_evm_commit -X github.com/ava-labs/subnet_evm/plugin/evm.Version=$subnet_evm_version" \
  -o $AVALANCHEGO_PLUGIN_DIR/srEXiWaHuhNyGwPUi444Tu47ZEDwxTWrbQiuD7FmgSAQ6X7Dy \
  "plugin/"*.go


export CHAIN_ID=99999
echo "creating genesis"

  cat <<EOF >$BASEDIR/genesis.json
{
  "config": {
    "chainId": $CHAIN_ID,
    "homesteadBlock": 0,
    "eip150Block": 0,
    "eip150Hash": "0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0",
    "eip155Block": 0,
    "eip158Block": 0,
    "byzantiumBlock": 0,
    "constantinopleBlock": 0,
    "petersburgBlock": 0,
    "istanbulBlock": 0,
    "muirGlacierBlock": 0,
    "subnetEVMTimestamp": 0,
    "feeConfig": {
      "gasLimit": 8000000,
      "minBaseFee": 25000000,
      "targetGas": 15000000,
      "baseFeeChangeDenominator": 36,
      "minBlockGasCost": 0,
      "maxBlockGasCost": 1000000,
      "targetBlockRate": 2,
      "blockGasCostStep": 20000
    },
    "contractDeployerAllowListConfig": {
      "blockTimestamp": 0,
      "adminAddresses": ["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"]
    },
    "orderBookConfig": {
      "blockTimestamp": 0,
      "adminAddresses": ["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"]
    },
    "contractNativeMinterConfig": {
      "blockTimestamp": 0,
      "adminAddresses": ["0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC"]
    },
    "allowFeeRecipients": false
  },
  "alloc": {
    "8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC": {
      "balance": "0xC9F2C9CD04674EDEA40000000"
    },
    "0Fa8EA536Be85F32724D57A37758761B86416123": {
      "balance": "0xC9F2C9CD04674EDEA40000000"
    }
  },
  "nonce": "0x0",
  "timestamp": "0x0",
  "extraData": "0x00",
  "gasLimit": "0x7A1200",
  "difficulty": "0x0",
  "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
  "coinbase": "0x0000000000000000000000000000000000000000",
  "number": "0x0",
  "gasUsed": "0x0",
  "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
}
EOF

#################################
# download avalanche-network-runner
# https://github.com/ava-labs/avalanche-network-runner
ANR_REPO_PATH=github.com/ava-labs/avalanche-network-runner
# version set
go install -v ${ANR_REPO_PATH}@${ANR_VERSION}

#################################
# run "avalanche-network-runner" server
GOPATH=$(go env GOPATH)
if [[ -z ${GOBIN+x} ]]; then
  # no gobin set
  BIN=${GOPATH}/bin/avalanche-network-runner
else
  # gobin set
  BIN=${GOBIN}/avalanche-network-runner
fi
echo "launch avalanche-network-runner in the background"
$BIN server \
  --log-level debug \
  --port=":12342" \
  --grpc-gateway-port=":12343" &
PID=${!}

$BIN control start \
  --log-level debug \
  --endpoint="0.0.0.0:12342" \
  --number-of-nodes=5 \
  --dial-timeout 30s \
  --avalanchego-path ${AVALANCHEGO_PATH} \
  --plugin-dir ${AVALANCHEGO_PLUGIN_DIR} \
  --blockchain-specs '[{"vm_name": "subnetevm", "genesis": "/tmp/subnet-evm-runner/genesis.json"}]'


echo "pkill -P ${PID} && kill -2 ${PID} && pkill -9 -f srEXiWaHuhNyGwPUi444Tu47ZEDwxTWrbQiuD7FmgSAQ6X7Dy" > kill.sh