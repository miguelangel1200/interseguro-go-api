#!/usr/bin/env bash
# Despliega interseguro-go-api en Cloud Run vía Cloud Build.
#
# Requisitos: gcloud autenticado y proyecto activo.
# Uso:        ./deploy.sh [PROJECT_ID] [REGION]
set -euo pipefail

PROJECT_ID="${1:-$(gcloud config get-value project 2>/dev/null)}"
REGION="${2:-us-central1}"

if [[ -z "$PROJECT_ID" ]]; then
  echo "ERROR: no se pudo determinar el PROJECT_ID." >&2
  echo "Uso: ./deploy.sh [PROJECT_ID] [REGION]" >&2
  exit 1
fi

IMAGE="${REGION}-docker.pkg.dev/${PROJECT_ID}/interseguro-go-api/interseguro-go-api:latest"

gcloud config set project "$PROJECT_ID"

echo "Subiendo imagen a Artifact Registry: $IMAGE"
gcloud builds submit --tag "$IMAGE" .
