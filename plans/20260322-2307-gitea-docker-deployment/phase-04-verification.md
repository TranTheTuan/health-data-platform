# Phase 4: Verification & Handoff

## Goal
Verify the deployment by triggering a tag push and checking the results on the target VM.

## Tasks
1. Commit and Push the `Dockerfile` and Gitea Action.
2. Tag a new version and push to `homelab`.
3. Check Gitea Actions output.
4. Verify the container is running and accessible.

---

### Step 1: Commit and Push CI/CD Config
Push the new Docker configuration and GitHub/Gitea Actions folder.

```bash
git add Dockerfile .gitea/workflows/deploy.yml
git commit -m "ci: add dockerization and gitea deployment workflow"
git push homelab main
```

### Step 2: Trigger a Tagged Release
Push a new version tag to trigger the CD.

```bash
git tag v1.0.1
git push homelab v1.0.1
```

### Step 3: Check Gitea Actions
Monitor the `Actions` tab in your Gitea repository.
- Ensure the `build-and-push` job completes successfully.
- Ensure the `deploy` job finishes without SSH errors.

### Step 4: Verify on the Target VM
Connect to your VM and check the container status:

```bash
cd ~/homelab/hdp
docker compose ps
docker logs -f hdp-api
```

---

## Troubleshooting Guide
- **SSH Error**: Ensure the private key is exactly as generated and the public key is in the VM's `authorized_keys`.
- **Registry Error**: Ensure Gitea Registry is enabled in Gitea's `app.ini` and public access/auth is configured.
- **Port Conflicts**: Ensure ports 8080 and 9090 are available on the target host.
