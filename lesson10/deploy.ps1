# Kubernetes Deployment Script for Bookstore API
# Usage: .\deploy.ps1 [-SkipBuild] [-Clean]

param(
    [switch]$SkipBuild,  # Skip Docker image builds
    [switch]$Clean       # Clean up existing deployment before deploying
)

$ErrorActionPreference = "Stop"

# Function to check if command succeeded
function Test-LastCommand {
    param([string]$Message)
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: $Message" -ForegroundColor Red
        exit 1
    }
}

# Clean up existing deployment if requested
if ($Clean) {
    Write-Host "`nCleaning up existing deployment..." -ForegroundColor Yellow
    kubectl delete namespace bookstore --ignore-not-found=true
    Write-Host "Waiting for namespace to be deleted..." -ForegroundColor Yellow
    Start-Sleep -Seconds 5
}

# Build Docker images
if (-not $SkipBuild) {
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host "STEP 1: Building Docker Images" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan

    Write-Host "`nBuilding migration image..." -ForegroundColor Green
    docker build -t bookstore-migrate:latest -f deployment/Dockerfile.migrate .
    Test-LastCommand "Failed to build migration image"

    Write-Host "`nBuilding API image..." -ForegroundColor Green
    docker build -t bookstore-api:latest -f Dockerfile .
    Test-LastCommand "Failed to build API image"
} else {
    Write-Host "`nSkipping Docker image builds..." -ForegroundColor Yellow
}

# Deploy to Kubernetes
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "STEP 2: Deploying to Kubernetes" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`n[1/10] Creating namespace..." -ForegroundColor Green
kubectl apply -f k8s/namespace.yaml
Test-LastCommand "Failed to create namespace"

Write-Host "`n[2/10] Creating secrets..." -ForegroundColor Green
kubectl apply -f k8s/postgres-secret.yaml
Test-LastCommand "Failed to create secrets"

Write-Host "`n[3/10] Creating persistent volume claim..." -ForegroundColor Green
kubectl apply -f k8s/postgres-pvc.yaml
Test-LastCommand "Failed to create PVC"

Write-Host "`n[4/10] Deploying PostgreSQL..." -ForegroundColor Green
kubectl apply -f k8s/postgres-deployment.yaml
Test-LastCommand "Failed to deploy PostgreSQL"

Write-Host "`n[5/10] Creating PostgreSQL service..." -ForegroundColor Green
kubectl apply -f k8s/postgres-service.yaml
Test-LastCommand "Failed to create PostgreSQL service"

Write-Host "`n[6/10] Waiting for PostgreSQL to be ready..." -ForegroundColor Yellow
kubectl wait --for=condition=ready pod -l app=postgres -n bookstore --timeout=300s
Test-LastCommand "PostgreSQL failed to become ready"

Write-Host "`n[7/10] Running database migrations..." -ForegroundColor Green
kubectl apply -f k8s/migrate-job.yaml
Test-LastCommand "Failed to create migration job"

Write-Host "`n[8/10] Waiting for migration to complete..." -ForegroundColor Yellow
kubectl wait --for=condition=complete job/migrate -n bookstore --timeout=300s
if ($LASTEXITCODE -ne 0) {
    Write-Host "Migration failed! Checking logs..." -ForegroundColor Red
    kubectl logs job/migrate -n bookstore
    exit 1
}

Write-Host "`n[9/10] Deploying API..." -ForegroundColor Green
kubectl apply -f k8s/api-deployment.yaml
Test-LastCommand "Failed to deploy API"

Write-Host "`n[10/10] Creating API service..." -ForegroundColor Green
kubectl apply -f k8s/api-service.yaml
Test-LastCommand "Failed to create API service"

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "STEP 3: Verifying Deployment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host "`nWaiting for API pods to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 5
kubectl wait --for=condition=ready pod -l app=bookstore-api -n bookstore --timeout=60s

Write-Host "`n✅ Deployment complete!" -ForegroundColor Green
Write-Host "`nCurrent status:" -ForegroundColor Cyan
kubectl get all -n bookstore

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Next Steps" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "`nAccess your API:" -ForegroundColor Yellow
Write-Host "  kubectl port-forward svc/bookstore-api-service 3000:3000 -n bookstore" -ForegroundColor White
Write-Host "`nThen visit:" -ForegroundColor Yellow
Write-Host "  • API:     http://localhost:3000" -ForegroundColor White
Write-Host "  • Swagger: http://localhost:3000/swagger/" -ForegroundColor White
Write-Host "  • Books:   http://localhost:3000/books" -ForegroundColor White

Write-Host "`nUseful commands:" -ForegroundColor Yellow
Write-Host "  • API logs:       kubectl logs -l app=bookstore-api -n bookstore" -ForegroundColor White
Write-Host "  • Migration logs: kubectl logs job/migrate -n bookstore" -ForegroundColor White
Write-Host "  • All resources:  kubectl get all -n bookstore" -ForegroundColor White
Write-Host "  • Delete all:     kubectl delete namespace bookstore" -ForegroundColor White
Write-Host ""
