#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        if [[ $key =~ ^#.*$ ]] || [[ -z $key ]]; then
            continue
        fi
        # Remove leading/trailing whitespace and quotes
        key=$(echo "$key" | xargs)
        value=$(echo "$value" | xargs)
        # Remove quotes if present
        value="${value%\"}"
        value="${value#\"}"
        export "$key=$value"
    done < .env
fi

cd cmd && go run *.go