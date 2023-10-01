# syntax=docker/dockerfile:experimental

# ============= Setting up base Stage ================
# Set required AVALANCHE_VERSION parameter in build image script
ARG AVALANCHE_VERSION

# ============= Compilation Stage ================
FROM golang:1.20.8-bullseye AS builder

WORKDIR /build

# Copy avalanche dependencies first (intermediate docker image caching)
# Copy avalanchego directory if present (for manual CI case, which uses local dependency)
COPY go.mod go.sum avalanchego* ./

# Download avalanche dependencies using go mod
RUN go mod download && go mod tidy -compat=1.20

# Copy the code into the container
COPY . .

# Pass in SUBNET_EVM_COMMIT as an arg to allow the build script to set this externally
ARG SUBNET_EVM_COMMIT
ARG CURRENT_BRANCH

RUN export SUBNET_EVM_COMMIT=$SUBNET_EVM_COMMIT && export CURRENT_BRANCH=$CURRENT_BRANCH && ./scripts/build.sh /build/o1Fg94YukvVRijwyThAavybVfwVJH3dhyz94g6qYRGdQ5Arqp

# ============= Cleanup Stage ================
FROM avaplatform/avalanchego:$AVALANCHE_VERSION AS builtImage

# Copy the evm binary into the correct location in the container
COPY --from=builder /build/o1Fg94YukvVRijwyThAavybVfwVJH3dhyz94g6qYRGdQ5Arqp /avalanchego/build/plugins/o1Fg94YukvVRijwyThAavybVfwVJH3dhyz94g6qYRGdQ5Arqp
