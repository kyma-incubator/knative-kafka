#!/usr/bin/env bash

# Ensure The Script Is Being Sourced
[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "\nWARNING: 'Source' this script for the environment setup to apply to the current session!\n"

# Absolute Path Of Directory Containing Script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# Set Local Testing Environment Variables
export HTTP_PORT="8080"
export METRICS_PORT="8081"
export KAFKA_BROKERS="localhost:9092"
export KAFKA_TOPIC="test-topic"
export CLIENT_ID="test-client-id"

# Log Environment Variables
echo ""
echo "Exported Env Vars..."
echo "    GOPATH=${GOPATH}"
echo "    HTTP_PORT=${HTTP_PORT}"
echo "    METRICS_PORT=${METRICS_PORT}"
echo "    KAFKA_BROKERS=${KAFKA_BROKERS}"
echo "    KAFKA_TOPIC=${KAFKA_TOPIC}"
echo "    CLIENT_ID=${CLIENT_ID}"
echo ""
