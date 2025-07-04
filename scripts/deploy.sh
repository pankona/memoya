#!/bin/bash

# Memoya Cloud Run Deployment Script
set -e

# Configuration
PROJECT_ID=${PROJECT_ID:-"memoya-mamemame"}
REGION=${REGION:-"asia-northeast1"}
SERVICE_NAME=${SERVICE_NAME:-"memoya-server"}
SERVICE_ACCOUNT=${SERVICE_ACCOUNT:-"memoya-server@${PROJECT_ID}.iam.gserviceaccount.com"}

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
    
    if ! command -v docker &> /dev/null; then
        echo_error "Docker is not installed. Please install it first."
        exit 1
    fi
    
    # Check if authenticated
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q "."; then
        echo_error "Not authenticated with gcloud. Please run 'gcloud auth login' first."
        exit 1
    fi
    
    echo_info "Prerequisites check passed."
}

# Function to set project
set_project() {
    echo_info "Setting project to $PROJECT_ID..."
    gcloud config set project $PROJECT_ID
    
    # Enable required APIs
    echo_info "Enabling required APIs..."
    gcloud services enable cloudbuild.googleapis.com
    gcloud services enable run.googleapis.com
    gcloud services enable containerregistry.googleapis.com
}

# Function to create service account if it doesn't exist
setup_service_account() {
    echo_info "Setting up service account..."
    
    # Check if service account exists
    if ! gcloud iam service-accounts describe $SERVICE_ACCOUNT &> /dev/null; then
        echo_info "Creating service account: $SERVICE_ACCOUNT"
        gcloud iam service-accounts create memoya-server \
            --display-name="Memoya Server" \
            --description="Service account for Memoya Cloud Run service"
    else
        echo_info "Service account $SERVICE_ACCOUNT already exists."
    fi
    
    # Grant necessary roles
    echo_info "Granting necessary roles to service account..."
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member="serviceAccount:$SERVICE_ACCOUNT" \
        --role="roles/firestore.user"
    
    gcloud projects add-iam-policy-binding $PROJECT_ID \
        --member="serviceAccount:$SERVICE_ACCOUNT" \
        --role="roles/secretmanager.secretAccessor"
}

# Function to build and deploy using Cloud Build
deploy_with_cloud_build() {
    echo_info "Building and deploying with Cloud Build..."
    gcloud builds submit --config cloudbuild.yaml .
}

# Function to deploy manually (alternative method)
deploy_manual() {
    echo_info "Building Docker image locally..."
    docker build -t gcr.io/$PROJECT_ID/$SERVICE_NAME:latest .
    
    echo_info "Pushing Docker image to Container Registry..."
    docker push gcr.io/$PROJECT_ID/$SERVICE_NAME:latest
    
    echo_info "Deploying to Cloud Run..."
    gcloud run deploy $SERVICE_NAME \
        --image gcr.io/$PROJECT_ID/$SERVICE_NAME:latest \
        --region $REGION \
        --platform managed \
        --allow-unauthenticated \
        --port 8080 \
        --memory 512Mi \
        --cpu 1 \
        --min-instances 0 \
        --max-instances 10 \
        --concurrency 80 \
        --timeout 300 \
        --set-env-vars PROJECT_ID=$PROJECT_ID \
        --service-account $SERVICE_ACCOUNT
}

# Function to get service URL
get_service_url() {
    echo_info "Getting service URL..."
    SERVICE_URL=$(gcloud run services describe $SERVICE_NAME \
        --region $REGION \
        --format "value(status.url)")
    
    echo_info "Service deployed successfully!"
    echo_info "Service URL: $SERVICE_URL"
    echo_info "Health check: $SERVICE_URL/health"
}

# Main deployment function
main() {
    echo_info "Starting Memoya Cloud Run deployment..."
    echo_info "Project: $PROJECT_ID"
    echo_info "Region: $REGION"
    echo_info "Service: $SERVICE_NAME"
    echo ""
    
    check_prerequisites
    set_project
    setup_service_account
    
    # Choose deployment method
    if [ "$1" = "--manual" ]; then
        deploy_manual
    else
        deploy_with_cloud_build
    fi
    
    get_service_url
    
    echo_info "Deployment completed successfully!"
}

# Show usage
usage() {
    echo "Usage: $0 [--manual]"
    echo ""
    echo "Options:"
    echo "  --manual    Deploy manually instead of using Cloud Build"
    echo ""
    echo "Environment variables:"
    echo "  PROJECT_ID      GCP project ID (default: memoya-mamemame)"
    echo "  REGION          Cloud Run region (default: asia-northeast1)"
    echo "  SERVICE_NAME    Cloud Run service name (default: memoya-server)"
    echo "  SERVICE_ACCOUNT Service account email"
    echo ""
    echo "Example:"
    echo "  PROJECT_ID=my-project ./scripts/deploy.sh"
    echo "  ./scripts/deploy.sh --manual"
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