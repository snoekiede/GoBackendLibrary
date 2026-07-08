# Kubernetes Deployment Order

## 🚀 Automated Deployment (Recommended)

```powershell
# Full deployment with image builds
.\deploy.ps1

# Skip image builds (use existing images)
.\deploy.ps1 -SkipBuild

# Clean up existing deployment first, then deploy
.\deploy.ps1 -Clean
```

---

## 📋 Manual Deployment Order

### Step 1: Build Docker Images

```powershell
# 1. Build migration image (includes migration SQL files)
docker build -t bookstore-migrate:latest -f deployment/Dockerfile.migrate .

# 2. Build API image
docker build -t bookstore-api:latest -f Dockerfile .
```

### Step 2: Deploy to Kubernetes

```powershell
# 3. Create namespace
kubectl apply -f k8s/namespace.yaml

# 4. Create secrets
kubectl apply -f k8s/postgres-secret.yaml

# 5. Create persistent volume claim
kubectl apply -f k8s/postgres-pvc.yaml

# 6. Deploy PostgreSQL
kubectl apply -f k8s/postgres-deployment.yaml

# 7. Create PostgreSQL service
kubectl apply -f k8s/postgres-service.yaml

# 8. Wait for PostgreSQL to be ready (IMPORTANT!)
kubectl wait --for=condition=ready pod -l app=postgres -n bookstore --timeout=60s

# 9. Run database migrations
kubectl apply -f k8s/migrate-job.yaml

# 10. Wait for migration to complete (IMPORTANT!)
kubectl wait --for=condition=complete job/migrate -n bookstore --timeout=120s

# 11. Deploy API application
kubectl apply -f k8s/api-deployment.yaml

# 12. Create API service
kubectl apply -f k8s/api-service.yaml
```

### Step 3: Verify Deployment

```powershell
# Check all resources
kubectl get all -n bookstore

# Check API logs
kubectl logs -l app=bookstore-api -n bookstore

# Check migration logs
kubectl logs job/migrate -n bookstore
```

### Step 4: Access the API

```powershell
# Port forward to access locally
kubectl port-forward svc/bookstore-api-service 3000:3000 -n bookstore
```

Then visit:
- **API**: http://localhost:3000
- **Swagger**: http://localhost:3000/swagger/
- **Books**: http://localhost:3000/books

---

## 📊 Deployment Flow Diagram

```
┌─────────────────────────────────────────────┐
│   Step 1: Build Docker Images               │
│   • bookstore-migrate:latest                │
│   • bookstore-api:latest                    │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Step 2-7: Setup Infrastructure            │
│   • Namespace (bookstore)                   │
│   • Secrets (database credentials)          │
│   • PVC (5GB storage)                       │
│   • PostgreSQL Deployment                   │
│   • PostgreSQL Service                      │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Step 8: Wait for PostgreSQL               │
│   ⏳ Polling until database is ready        │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Step 9-10: Run Migrations                 │
│   • Create tables (books, users, etc.)      │
│   • Apply schema changes                    │
│   ⏳ Wait for completion                    │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   Step 11-12: Deploy API                    │
│   • API Deployment (3 replicas)             │
│   • API Service (LoadBalancer)              │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│   ✅ Deployment Complete!                   │
│   Ready to accept requests                  │
└─────────────────────────────────────────────┘
```

---

## ⚠️ Important Notes

### Order Matters!
1. **PostgreSQL must be ready** before running migrations
2. **Migrations must complete** before starting the API
3. The script includes proper wait commands to ensure this

### Wait Commands
- `kubectl wait --for=condition=ready` - Waits for pods to be healthy
- `kubectl wait --for=condition=complete` - Waits for jobs to finish

### Why This Order?
- **Secrets before deployments**: Pods need credentials on startup
- **PVC before database**: Database needs storage available
- **Database before migrations**: Can't migrate a non-existent database
- **Migrations before API**: API expects tables to exist

---

## 🔄 Updating After Initial Deployment

### Update API Only
```powershell
# Rebuild image
docker build -t bookstore-api:latest -f Dockerfile .

# Restart deployment
kubectl rollout restart deployment/bookstore-api -n bookstore
```

### Update Migrations
```powershell
# Rebuild migration image
docker build -t bookstore-migrate:latest -f deployment/Dockerfile.migrate .

# Delete old job and rerun
kubectl delete job migrate -n bookstore
kubectl apply -f k8s/migrate-job.yaml
```

### Full Redeployment
```powershell
# Clean up and redeploy
.\deploy.ps1 -Clean
```

---

## 🧹 Cleanup

```powershell
# Delete everything
kubectl delete namespace bookstore

# This removes:
# ✓ All pods
# ✓ All deployments
# ✓ All services
# ✓ All secrets
# ✓ All jobs
# ✓ Persistent volume claims and data
```

---

## 🎯 Quick Commands Reference

```powershell
# Deploy
.\deploy.ps1

# Check status
kubectl get all -n bookstore

# View logs
kubectl logs -l app=bookstore-api -n bookstore -f

# Access API
kubectl port-forward svc/bookstore-api-service 3000:3000 -n bookstore

# Restart API
kubectl rollout restart deployment/bookstore-api -n bookstore

# Delete all
kubectl delete namespace bookstore
```

