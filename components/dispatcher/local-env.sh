#!/usr/bin/env bash

# Ensure The Script Is Being Sourced
[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "\nWARNING: 'Source' this script for the environment setup to apply to the current session!\n"

# Absolute Path Of Directory Containing Script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"


# Set Local Testing Environment Variables
export KAFKA_GROUP_ID="test-group-id"
export METRICS_PORT="8081"
export KAFKA_BROKERS="localhost:9092"
export KAFKA_TOPIC="test-topic"
export KAFKA_CONSUMERS=4
export KAFKA_OFFSET="Newest"
export SUBSCRIBER_URI="test-subscriber-uri"
export EXPONENTIAL_BACKOFF=true
export INITIAL_RETRY_INTERVAL=500
export MAX_RETRY_TIME=30000


# Log Environment Variables
echo ""
echo "Exported Env Vars..."
echo "    GOPATH=${GOPATH}"
echo "    METRICS_PORT=${METRICS_PORT}"
echo "    KAFKA_GROUP_ID=${KAFKA_GROUP_ID}"
echo "    KAFKA_BROKERS=${KAFKA_BROKERS}"
echo "    KAFKA_TOPIC=${KAFKA_TOPIC}"
echo "    KAFKA_CONSUMERS=${KAFKA_CONSUMERS}"
echo "    KAFKA_OFFSET=${KAFKA_OFFSET}"
echo "    SUBSCRIBER_URI=${SUBSCRIBER_URI}"
echo "    EXPONENTIAL_BACKOFF=${EXPONENTIAL_BACKOFF}"
echo "    INITIAL_RETRY_INTERVAL=${INITIAL_RETRY_INTERVAL}"
echo "    MAX_RETRY_TIME=${MAX_RETRY_TIME}"
echo ""
