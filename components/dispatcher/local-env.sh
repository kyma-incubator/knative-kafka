#!/usr/bin/env bash

# Ensure The Script Is Being Sourced
[[ "${BASH_SOURCE[0]}" == "${0}" ]] && echo -e "\nWARNING: 'Source' this script for the environment setup to apply to the current session!\n"

# Absolute Path Of Directory Containing Script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"


# Set Local Testing Environment Variables
export METRICS_PORT="8081"
export KAFKA_BROKERS="localhost:9092"
export KAFKA_USERNAME=""
export KAFKA_PASSWORD=""
export KAFKA_TOPIC="test-topic"
export KAFKA_CONSUMERS=4
export KAFKA_OFFSET="Newest"
export EXPONENTIAL_BACKOFF=true
export INITIAL_RETRY_INTERVAL=500
export MAX_RETRY_TIME=30000


# Log Environment Variables
echo ""
echo "Exported Env Vars..."
echo ""
echo "GOPATH=${GOPATH}"
echo "METRICS_PORT=${METRICS_PORT}"
echo "CHANNEL_KEY=${CHANNEL_KEY}"
echo "KAFKA_BROKERS=${KAFKA_BROKERS}"
echo "KAFKA_USERNAME=${KAFKA_USERNAME}"
echo "KAFKA_PASSWORD=${KAFKA_PASSWORD}"
echo "KAFKA_TOPIC=${KAFKA_TOPIC}"
echo "KAFKA_CONSUMERS=${KAFKA_CONSUMERS}"
echo "KAFKA_OFFSET=${KAFKA_OFFSET}"
echo "EXPONENTIAL_BACKOFF=${EXPONENTIAL_BACKOFF}"
echo "KAFKA_OFFSET_COMMIT_MESSAGE_COUNT=${KAFKA_OFFSET_COMMIT_MESSAGE_COUNT}"
echo "KAFKA_OFFSET_COMMIT_DURATION_MILLIS=${KAFKA_OFFSET_COMMIT_DURATION_MILLIS}"
echo "INITIAL_RETRY_INTERVAL=${INITIAL_RETRY_INTERVAL}"
echo "MAX_RETRY_TIME=${MAX_RETRY_TIME}"
echo ""
