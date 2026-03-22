# Phase 2: Gitea Action Workflow

## Goal
Automate the build, push, and deployment using Gitea Actions.

## Tasks
Create `.gitea/workflows/deploy.yml` with the following content.

### Step 1: Workflow Configuration (`.gitea/workflows/deploy.yml`)
Trigger the build and push based on git tags matching `v*`.

```yaml
name: Gitea Docker CI/CD (Local LAN Only)

on:
  push:
    tags:
      - 'v*'

jobs:
  build-and-push:
    runs-on: self-hosted  # Use the local runner on the GAN LAN
    steps:
      - name: Checkout Source
        uses: actions/checkout@v4

      - name: Log in to Gitea Container Registry
        run: echo "${{ secrets.GITEA_TOKEN }}" | docker login 192.168.1.222:3000 -u tuantt --password-stdin
        
      - name: Build Docker Image
        run: |
          docker build -t 192.168.1.222:3000/tuantt/hdp:${{ github.ref_name }} .
          docker tag 192.168.1.222:3000/tuantt/hdp:${{ github.ref_name }} 192.168.1.222:3000/tuantt/hdp:latest

      - name: Push Docker Image
        run: |
          docker push 192.168.1.222:3000/tuantt/hdp:${{ github.ref_name }}
          docker push 192.168.1.222:3000/tuantt/hdp:latest

  deploy:
    needs: build-and-push
    runs-on: self-hosted
    steps:
      - name: SSH Deploy to VM
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          port: 22
          script: |
              cd ~/homelab/hdp
              docker compose pull
              docker compose up -d
              docker image prune -f
```

---

## Secret Configuration in Gitea Settings
Add the following secrets to your Gitea repository at `Settings > Actions > Secrets`:
1. `GITEA_TOKEN`: Personal Access Token from Gitea (with `write:packages` and `read:packages` flags).
2. `SSH_HOST`: IP address of the target VM (`192.168.1.222` or as appropriate).
3. `SSH_USER`: The user account with docker permissions on the VM.
4. `SSH_PRIVATE_KEY`: The private key associated with the `hdp_lab.pub` earlier.
