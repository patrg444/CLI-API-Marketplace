#!/bin/bash

# Validate Go syntax for new commands
echo "🔍 Validating Go syntax for marketplace commands..."

# Run in Docker to check syntax
docker run --rm \
  -v "$(pwd)":/app \
  -w /app \
  golang:1.21-alpine \
  sh -c '
    echo "Checking individual command files..."
    
    files=(
      "cmd/analytics.go"
      "cmd/earnings.go"
      "cmd/subscriptions.go"
      "cmd/review.go"
      "cmd/search.go"
    )
    
    for file in "${files[@]}"; do
      echo -n "Checking $file... "
      if go build -o /dev/null "$file" 2>/tmp/error.log; then
        echo "✅"
      else
        echo "❌"
        cat /tmp/error.log
      fi
    done
  '