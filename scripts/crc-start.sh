#!/usr/bin/env bash

set -e

# This script manages the creation of the CRC cluster to be used for testing
# Usage: crc-start.sh
# FORCE_CLEAN=1 crc-start.sh will delete any current crc cluster, clear the cache and start a fresh installation
# If CRC is already running, nothing happens

# Check CRC is installed
if ! [ -x "$(command -v crc)" ]; then
  echo "Error: CRC is not installed. Go to: https://developers.redhat.com/products/codeready-containers/overview"
  exit 1
fi

# Blast CRC if necessary
if [ "${FORCE_CLEAN}" = 1 ]; then
  crc delete --clear-cache --force
fi

# Start CRC if necessary
if [ "$(crc status -o json | jq -r .success)" = "false"  ] || [ "$(crc status -o json | jq -r .crcStatus)" = "Stopped" ]; then
    echo "Setting up CRC"
    crc setup
    crc start
fi

# Check CRC started successfully
if ! [ "$(crc status -o json | jq -r .crcStatus)" = "Running" ]; then
  echo "Error: CRC is unreachable. Please try recreating the cluster."
  exit 1
fi

echo "SUCCESS! kubeconfig=${HOME}/.crc/machines/crc/kubeconfig"
