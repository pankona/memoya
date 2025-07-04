#!/bin/bash

# Memoya GCP Setup Script
# This script sets up everything needed for Cloud Run deployment via gcloud CLI
set -e

# Configuration
PROJECT_ID=${PROJECT_ID:-"memoya-mamemame"}
REGION=${REGION:-"asia-northeast1"}
SERVICE_ACCOUNT_NAME="memoya-server"
SERVICE_ACCOUNT_EMAIL="$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

echo_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Function to check if required tools are installed
check_prerequisites() {
    echo_step "Checking prerequisites..."
    
    if ! command -v gcloud &> /dev/null; then
        echo_error "gcloud CLI is not installed. Please install it first."
        echo "  https://cloud.google.com/sdk/docs/install"
        exit 1
    fi
    
    # Check if authenticated
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "."; then
        echo_error "Not authenticated with gcloud. Please run 'gcloud auth login' first."
        exit 1
    fi
    
    echo_info "Prerequisites check passed."
}

# Function to set project and enable APIs
setup_project() {
    echo_step "Setting up project and enabling APIs..."
    
    # Set project
    echo_info "Setting project to $PROJECT_ID..."
    gcloud config set project $PROJECT_ID
    
    # Enable required APIs
    echo_info "Enabling required APIs..."
    local apis=(
        "cloudbuild.googleapis.com"
        "run.googleapis.com"
        "containerregistry.googleapis.com"
        "secretmanager.googleapis.com"
        "firestore.googleapis.com"
        "firebase.googleapis.com"
    )
    
    for api in "${apis[@]}"; do
        echo_info "  Enabling $api..."
        gcloud services enable $api
    done
    
    echo_info "APIs enabled successfully."
}

# Function to create Firestore database
setup_firestore() {
    echo_step "Setting up Firestore database..."
    
    # Check if Firestore database already exists
    if gcloud firestore databases describe --region=$REGION &> /dev/null; then
        echo_info "Firestore database already exists."
        return 0
    fi
    
    echo_info "Creating Firestore database in region $REGION..."
    
    # Try to create Firestore database
    if gcloud firestore databases create --region=$REGION 2>/dev/null; then
        echo_info "Firestore database created successfully."
    else
        echo_warn "Failed to create Firestore database via gcloud."
        echo_warn "Please create it manually:"
        echo_warn "  1. Go to Firebase Console: https://console.firebase.google.com/"
        echo_warn "  2. Select project: $PROJECT_ID"
        echo_warn "  3. Go to Firestore Database > Create Database"
        echo_warn "  4. Choose 'Start in production mode' or 'Start in test mode'"
        echo_warn "  5. Select location: $REGION"
        echo ""
        read -p "Press Enter after creating Firestore database manually..."
    fi
}

# Function to create service account
setup_service_account() {
    echo_step "Setting up service account..."
    
    # Check if service account exists
    if gcloud iam service-accounts describe $SERVICE_ACCOUNT_EMAIL &> /dev/null; then
        echo_info "Service account $SERVICE_ACCOUNT_EMAIL already exists."
    else
        echo_info "Creating service account: $SERVICE_ACCOUNT_NAME"
        gcloud iam service-accounts create $SERVICE_ACCOUNT_NAME \
            --display-name="Memoya Server" \
            --description="Service account for Memoya Cloud Run service"
    fi
    
    # Grant necessary roles
    echo_info "Granting necessary roles to service account..."
    local roles=(
        "roles/firestore.user"
        "roles/secretmanager.secretAccessor"
    )
    
    for role in "${roles[@]}"; do
        echo_info "  Granting $role..."
        gcloud projects add-iam-policy-binding $PROJECT_ID \
            --member="serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
            --role="$role" --quiet
    done
    
    echo_info "Service account setup completed."
}

# Function to setup Secret Manager
setup_secrets() {
    echo_step "Setting up Secret Manager..."
    
    local secrets=("oauth-client-id" "oauth-client-secret")
    
    for secret in "${secrets[@]}"; do
        if gcloud secrets describe $secret --project=$PROJECT_ID &> /dev/null; then
            echo_info "Secret $secret already exists."
        else
            echo_info "Creating secret: $secret"
            echo "REPLACE_WITH_ACTUAL_VALUE" | gcloud secrets create $secret \
                --replication-policy="automatic" \
                --data-file=- --project=$PROJECT_ID
        fi
        
        # Grant access to service account
        echo_info "Granting secret access to service account for $secret..."
        gcloud secrets add-iam-policy-binding $secret \
            --member="serviceAccount:$SERVICE_ACCOUNT_EMAIL" \
            --role="roles/secretmanager.secretAccessor" \
            --project=$PROJECT_ID --quiet
    done
    
    echo_info "Secret Manager setup completed."
}

# Function to show manual steps
show_manual_steps() {
    echo_step "Manual steps required..."
    echo ""
    echo_warn "The following steps must be completed manually in Google Cloud Console:"
    echo ""
    echo "1. 🔑 OAuth 2.0 Setup (REQUIRED):"
    echo "   URL: https://console.cloud.google.com/apis/credentials?project=$PROJECT_ID"
    echo "   Steps:"
    echo "     a) Click '+ CREATE CREDENTIALS' > 'OAuth 2.0 Client ID'"
    echo "     b) Choose 'Desktop application' as Application type"
    echo "     c) Set name: 'Memoya Desktop Client'"
    echo "     d) Click 'CREATE'"
    echo "     e) Copy Client ID and Client Secret"
    echo ""
    echo "2. 📝 Update Secret Manager with OAuth credentials:"
    echo "   After getting OAuth credentials, run:"
    echo "     echo 'YOUR_CLIENT_ID' | gcloud secrets versions add oauth-client-id --data-file=-"
    echo "     echo 'YOUR_CLIENT_SECRET' | gcloud secrets versions add oauth-client-secret --data-file=-"
    echo ""
    echo "3. 🗄️ Verify Firestore Database:"
    echo "   URL: https://console.firebase.google.com/project/$PROJECT_ID/firestore"
    echo "   Ensure database is created and accessible"
    echo ""
}

# Function to verify setup
verify_setup() {
    echo_step "Verifying setup..."
    
    local issues=()
    
    # Check APIs
    local required_apis=(
        "cloudbuild.googleapis.com"
        "run.googleapis.com"
        "secretmanager.googleapis.com"
        "firestore.googleapis.com"
    )
    
    for api in "${required_apis[@]}"; do
        if gcloud services list --enabled --filter="name:$api" --format="value(name)" | grep -q "$api"; then
            echo_info "✓ API enabled: $api"
        else
            issues+=("API not enabled: $api")
        fi
    done
    
    # Check service account
    if gcloud iam service-accounts describe $SERVICE_ACCOUNT_EMAIL &> /dev/null; then
        echo_info "✓ Service account exists: $SERVICE_ACCOUNT_EMAIL"
    else
        issues+=("Service account not found: $SERVICE_ACCOUNT_EMAIL")
    fi
    
    # Check secrets
    local secrets=("oauth-client-id" "oauth-client-secret")
    for secret in "${secrets[@]}"; do
        if gcloud secrets describe $secret --project=$PROJECT_ID &> /dev/null; then
            echo_info "✓ Secret exists: $secret"
        else
            issues+=("Secret not found: $secret")
        fi
    done
    
    # Check Firestore
    if gcloud firestore databases describe --region=$REGION &> /dev/null; then
        echo_info "✓ Firestore database exists"
    else
        issues+=("Firestore database not found")
    fi
    
    if [ ${#issues[@]} -eq 0 ]; then
        echo_info "All checks passed! ✨"
        return 0
    else
        echo_warn "Issues found:"
        for issue in "${issues[@]}"; do
            echo_warn "  ✗ $issue"
        done
        return 1
    fi
}

# Function to show next steps
show_next_steps() {
    echo_step "Next steps..."
    echo ""
    echo_info "1. Complete OAuth setup in Cloud Console (see manual steps above)"
    echo_info "2. Update OAuth secrets in Secret Manager"
    echo_info "3. Run deployment:"
    echo "     ./scripts/deploy.sh"
    echo ""
    echo_info "4. Or run setup-secrets.sh for interactive OAuth setup:"
    echo "     ./scripts/setup-secrets.sh"
    echo ""
}

# Main setup function
main() {
    echo_info "Starting Memoya GCP setup..."
    echo_info "Project: $PROJECT_ID"
    echo_info "Region: $REGION"
    echo_info "Service Account: $SERVICE_ACCOUNT_EMAIL"
    echo ""
    
    check_prerequisites
    setup_project
    setup_firestore
    setup_service_account
    setup_secrets
    
    echo ""
    show_manual_steps
    
    echo ""
    if verify_setup; then
        echo_info "GCP setup completed successfully! 🎉"
    else
        echo_warn "Setup completed with some issues. Please resolve them before deploying."
    fi
    
    show_next_steps
}

# Show usage
usage() {
    echo "Usage: $0"
    echo ""
    echo "Environment variables:"
    echo "  PROJECT_ID    GCP project ID (default: memoya-mamemame)"
    echo "  REGION        GCP region (default: asia-northeast1)"
    echo ""
    echo "This script will:"
    echo "  1. Enable required APIs"
    echo "  2. Create Firestore database"
    echo "  3. Create service account with necessary permissions"
    echo "  4. Setup Secret Manager secrets"
    echo "  5. Show manual steps for OAuth setup"
    echo ""
    echo "Prerequisites:"
    echo "  - gcloud CLI installed and authenticated"
    echo "  - Appropriate permissions in the GCP project"
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