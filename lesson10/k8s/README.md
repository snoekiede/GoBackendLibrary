# Kubernetes Deployment for Bookstore API

This directory contains Kubernetes manifest files for deploying the Bookstore API application.

## Architecture

The deployment consists of:
- **PostgreSQL Database**: Persistent database with health checks
- **Migration Job**: Runs database migrations using goose
- **API Application**: 3 replicas of the bookstore API with health checks
- **Services**: ClusterIP for database, LoadBalancer for API

## Prerequisites

- Kubernetes cluster (Docker Desktop, Minikube, or Kind)
- kubectl configured to access your cluster
- Docker for building images

## Quick Start

### 1. Build Docker Images

```powershell
# Build migration image (includes SQL migration files)
docker build -t bookstore-migrate:latest -f deployment/Dockerfile.migrate .

# Build API image
docker build -t bookstore-api:latest -f Dockerfile .
```

**For Minikube users:**
```powershell
# Use Minikube's Docker daemon
minikube docker-env | Invoke-Expression
# Then build images as shown above
```

**For Kind users:**
```powershell
# Load images into Kind cluster
kind load docker-image bookstore-migrate:latest
kind load docker-image bookstore-api:latest
```

### 2. Deploy to Kubernetes

#### Option A: Use the deployment script
```powershell
.\deploy.ps1
```

#### Option B: Manual deployment
```powershell
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Create secrets
kubectl apply -f k8s/postgres-secret.yaml

# Create persistent volume
kubectl apply -f k8s/postgres-pvc.yaml

# Deploy PostgreSQL
kubectl apply -f k8s/postgres-deployment.yaml
kubectl apply -f k8s/postgres-service.yaml

# Wait for PostgreSQL to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n bookstore --timeout=60s

# Run migrations
kubectl apply -f k8s/migrate-job.yaml

# Wait for migrations to complete
kubectl wait --for=condition=complete job/migrate -n bookstore --timeout=120s

# Deploy API
kubectl apply -f k8s/api-deployment.yaml
kubectl apply -f k8s/api-service.yaml
```

## Accessing the API

### Check deployment status
```powershell
kubectl get all -n bookstore
```

### View API logs
```powershell
kubectl logs -l app=bookstore-api -n bookstore
```

### View migration logs
```powershell
kubectl logs job/migrate -n bookstore
```

### Access the API
```powershell
# Get the service details
kubectl get svc bookstore-api-service -n bookstore

# Port forward to access locally
kubectl port-forward svc/bookstore-api-service 3000:3000 -n bookstore
```

Then access:
- API: http://localhost:3000
- Swagger docs: http://localhost:3000/swagger/

## Manifest Files

- `namespace.yaml` - Creates the bookstore namespace
- `postgres-secret.yaml` - Database credentials (change in production!)
- `postgres-pvc.yaml` - Persistent volume for database data
- `postgres-deployment.yaml` - PostgreSQL deployment with health checks
- `postgres-service.yaml` - ClusterIP service for database
- `migrate-job.yaml` - One-time job to run database migrations
- `api-deployment.yaml` - API deployment with 3 replicas and health checks
- `api-service.yaml` - LoadBalancer service for API

## Configuration

### Database Credentials
Edit `k8s/postgres-secret.yaml` to change database credentials. In production, use proper secret management (Azure Key Vault, AWS Secrets Manager, etc.)

### API Replicas
Edit `api-deployment.yaml` to change the number of replicas:
```yaml
spec:
  replicas: 3  # Change this value
```

### Storage
Edit `postgres-pvc.yaml` to change storage size:
```yaml
resources:
  requests:
    storage: 5Gi  # Change this value
```

## Troubleshooting

### Pods not starting
```powershell
kubectl describe pod <pod-name> -n bookstore
kubectl logs <pod-name> -n bookstore
```

### Migration failed
```powershell
# Check migration logs
kubectl logs job/migrate -n bookstore

# Re-run migration
kubectl delete job migrate -n bookstore
kubectl apply -f k8s/migrate-job.yaml
```

### Database connection issues
```powershell
# Check if PostgreSQL is running
kubectl get pods -l app=postgres -n bookstore

# Test database connection
kubectl exec -it <postgres-pod-name> -n bookstore -- psql -U postgres -d gobooksnew
```

## Clean Up

```powershell
# Delete all resources
kubectl delete namespace bookstore

# Or delete individual resources
kubectl delete -f k8s/
```

## Notes

- **Health Checks**: The API uses the root endpoint `/` for liveness and readiness probes
- **Local Images**: Configured to use local Docker images with `imagePullPolicy: Never`
- **Migration Files**: Built into the migration Docker image, no ConfigMaps needed
- **Database**: Data persists via PersistentVolumeClaim even if pods restart

## Production Considerations

For production deployments:
1. Use proper secret management (not plain YAML files)
2. Configure resource limits and requests
3. Set up proper monitoring and logging
4. Use Ingress instead of LoadBalancer
5. Configure backup strategy for PostgreSQL
6. Use image tags instead of `latest`
7. Implement network policies
8. Configure pod disruption budgets
9. Use horizontal pod autoscaling
10. Store images in a container registry

