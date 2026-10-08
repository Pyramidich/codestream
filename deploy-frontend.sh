#!/usr/bin/env bash
set -euo pipefail

# Deploy frontend to ITMO server
# Usage: ./deploy-frontend.sh <username>
# Example: ./deploy-frontend.sh s503255

USER=${1:-}

if [ -z "$USER" ]; then
  echo "Usage: $0 <username>"
  echo "Example: $0 s503255"
  exit 1
fi

HOST="se.ifmo.ru"
REMOTE_PATH="$USER@$HOST:~/public_html/codestream"

echo "Building frontend..."
cd "$(dirname "$0")/frontend"
npm run build

echo "Creating remote directory..."
ssh -p 2222 "$USER@$HOST" "mkdir -p ~/public_html/codestream"

echo "Uploading files..."
rsync -avz --delete -e "ssh -p 2222" dist/ "$REMOTE_PATH/"

echo "Deployed to https://se.ifmo.ru/~$USER/codestream/"
