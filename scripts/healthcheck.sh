#!/bin/sh

# Health check script for OnlyFans Event Publisher
# This script can be used in Docker health checks

set -e

# Check if the main process is running
if ! pgrep -f "./main" > /dev/null; then
    echo "Main process not running"
    exit 1
fi

echo "Health check passed"
exit 0