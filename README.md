# Passion-Tree Backend (Go)

## Dev: Run with Docker hot-reload

- Prereqs: Install Docker Desktop.
- This uses Air to rebuild and restart on code changes.

### Troubleshooting

**If you see `archive/tar: unknown file mode` error:**

Run this command once:
```bash
git config core.fileMode false
```

Then rebuild:
```bash
cd ../Passion-Tree-Infrastructure
docker compose down
./scripts/dev-up.sh  # or dev-up.ps1 on Windows
```

### Start

From the infrastructure folder:

```powershell
cd ..\Passion-Tree-Infrastructure
./scripts/dev-up.ps1 -Rebuild
```

- Backend listens on `http://localhost:8080`.
- Health check: `GET /health`.

### How it works

- `docker-compose.override.yml` overrides `backend-go` to build
	from `../Passion-Tree-Backend/Dockerfile.dev`, mounts the source,
	and runs `air` with `.air.toml`.
- Any saved `.go` or `go.mod` changes trigger rebuild + restart.

### Stop

In the same folder:

```powershell
docker compose -f docker-compose.yml -f docker-compose.override.yml down
```

## Notes

- Environment `PORT=8080` is set in Compose; adjust if needed.
- Go module name is `passiontree`; standard `go build` is used.

## Production: Azure Container Apps

- The Air-based setup is for development only; do not use it in production.
- Build a production image using the lean `Dockerfile` (no Air, non-root user).

### Build locally

```powershell
cd ..\Passion-Tree-Backend
docker build -t backend-go:prod -f Dockerfile .
```

### Push to Azure Container Registry (ACR)

Replace placeholders with your values.

```powershell
$Registry = "<yourRegistry>" # e.g. myregistry.azurecr.io
docker tag backend-go:prod $Registry/backend-go:latest
docker push $Registry/backend-go:latest
```

### Point Terraform to your image

Edit the image reference in the Container App resource to use your ACR image:

- See [Passion-Tree-Infrastructure/terraform/container_apps.tf](Passion-Tree-Infrastructure/terraform/container_apps.tf#L18-L29) and update `image` under `container` to `$Registry/backend-go:latest`.

Then apply:

```powershell
cd ..\Passion-Tree-Infrastructure\terraform
terraform init
terraform apply
```

### Recommended production settings

- Set environment variables for `DB_URL`, `AI_SERVICE_URL` in the `env` block of the container.
- Configure readiness/liveness probes hitting `/health` on `8080`.
- Use `revision_mode = "Single"` and traffic weights per release.
# Passion-Tree-Backend