#!/bin/bash

# Memoya Secret Manager Setup Script
set -e

# Configuration
PROJECT_ID=${PROJECT_ID:-"memoya-mamemame"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if required tools are installed
check_prerequisites() {
    echo_info "Checking prerequisites..."
    
    if ! command -v gcloud &> /dev/null; then
        echo_error "gcloud CLI is not installed. Please install it first."
        exit 1
    fi
    
    # Check if authenticated
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "."; then
        echo_error "Not authenticated with gcloud. Please run 'gcloud auth login' first."
        exit 1
    fi
    
    echo_info "Prerequisites check passed."
}

# Function to enable Secret Manager API
enable_apis() {
    echo_info "Enabling Secret Manager API..."
    gcloud services enable secretmanager.googleapis.com --project=$PROJECT_ID
}

# Function to create a secret
create_secret() {
    local secret_name=$1
    local secret_description=$2
    
    echo_info "Creating secret: $secret_name"
    
    # Check if secret already exists
    if gcloud secrets describe $secret_name --project=$PROJECT_ID &> /dev/null; then
        echo_warn "Secret $secret_name already exists. Skipping creation."
        return 0
    fi
    
    # Create the secret
    gcloud secrets create $secret_name \
        --replication-policy="automatic" \
        --project=$PROJECT_ID \
        --data-file=/dev/stdin <<< "REPLACE_WITH_ACTUAL_VALUE"
    
    echo_info "Secret $secret_name created successfully."
    echo_warn "Please update the secret value using: gcloud secrets versions add $secret_name --data-file=-"
}

# Function to grant access to service account
grant_secret_access() {
    local service_account="memoya-server@$PROJECT_ID.iam.gserviceaccount.com"
    local secrets=("oauth-client-id" "oauth-client-secret")
    
    echo_info "Granting secret access to service account: $service_account"
    
    for secret in "${secrets[@]}"; do
        echo_info "Granting access to secret: $secret"
        gcloud secrets add-iam-policy-binding $secret \
            --member="serviceAccount:$service_account" \
            --role="roles/secretmanager.secretAccessor" \
            --project=$PROJECT_ID
    done
}

# Function to set secret values
set_secret_values() {
    echo_info "Setting up OAuth secrets..."
    
    # Prompt for OAuth Client ID
    echo ""
    echo "Please enter your Google OAuth 2.0 Client ID:"
    echo "(You can get this from Google Cloud Console > APIs & Services > Credentials)"
    read -p "OAuth Client ID: " oauth_client_id
    
    if [ -n "$oauth_client_id" ]; then
        echo_info "Setting oauth-client-id secret..."
        echo "$oauth_client_id" | gcloud secrets versions add oauth-client-id \
            --data-file=- --project=$PROJECT_ID
        echo_info "OAuth Client ID secret updated successfully."
    else
        echo_warn "OAuth Client ID not provided. Please set it manually later."
    fi
    
    # Prompt for OAuth Client Secret
    echo ""
    echo "Please enter your Google OAuth 2.0 Client Secret:"
    read -s -p "OAuth Client Secret: " oauth_client_secret
    echo ""
    
    if [ -n "$oauth_client_secret" ]; then
        echo_info "Setting oauth-client-secret secret..."
        echo "$oauth_client_secret" | gcloud secrets versions add oauth-client-secret \
            --data-file=- --project=$PROJECT_ID
        echo_info "OAuth Client Secret secret updated successfully."
    else
        echo_warn "OAuth Client Secret not provided. Please set it manually later."
    fi
}

# Function to list secrets
list_secrets() {
    echo_info "Listing all secrets in project $PROJECT_ID:"
    gcloud secrets list --project=$PROJECT_ID --format="table(name,createTime,replication.automatic)"
}

# Function to test secret access
test_secret_access() {
    echo_info "Testing secret access..."
    
    local secrets=("oauth-client-id" "oauth-client-secret")
    
    for secret in "${secrets[@]}"; do
        echo_info "Testing access to secret: $secret"
        if gcloud secrets versions access latest --secret=$secret --project=$PROJECT_ID &> /dev/null; then
            echo_info "✓ Access to $secret: OK"
        else
            echo_error "✗ Access to $secret: FAILED"
        fi
    done
}

# Main setup function
main() {
    echo_info "Setting up Secret Manager for Memoya..."
    echo_info "Project: $PROJECT_ID"
    echo ""
    
    check_prerequisites
    
    # Set project
    gcloud config set project $PROJECT_ID
    
    enable_apis
    
    # Create secrets
    create_secret "oauth-client-id" "Google OAuth 2.0 Client ID for Memoya"
    create_secret "oauth-client-secret" "Google OAuth 2.0 Client Secret for Memoya"
    
    # Grant access to service account
    grant_secret_access
    
    # Ask if user wants to set secret values now
    echo ""
    read -p "Do you want to set the OAuth secret values now? (y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        set_secret_values
    else
        echo_warn "Secrets created with placeholder values. Please set them manually:"
        echo "  gcloud secrets versions add oauth-client-id --data-file=- --project=$PROJECT_ID"
        echo "  gcloud secrets versions add oauth-client-secret --data-file=- --project=$PROJECT_ID"
    fi
    
    echo ""
    list_secrets
    
    echo ""
    test_secret_access
    
    echo ""
    echo_info "Secret Manager setup completed successfully!"
    echo_info "Your OAuth credentials are now stored securely in Secret Manager."
    echo_info "The Cloud Run service will automatically access them using the service account."
}

# Show usage
usage() {
    echo "Usage: $0"
    echo ""
    echo "Environment variables:"
    echo "  PROJECT_ID    GCP project ID (default: memoya-mamemame)"
    echo ""
    echo "This script will:"
    echo "  1. Enable Secret Manager API"
    echo "  2. Create oauth-client-id and oauth-client-secret secrets"
    echo "  3. Grant access to the memoya-server service account"
    echo "  4. Optionally set the secret values"
    echo ""
    echo "Prerequisites:"
    echo "  - gcloud CLI installed and authenticated"
    echo "  - Appropriate permissions in the GCP project"
    echo "  - Service account memoya-server@PROJECT_ID.iam.gserviceaccount.com exists"
}

# Handle script arguments
case "$1" in
    -h|--help)
        usage
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac