# Phase 3: Server Environment Setup

## Goal
Prepare the on-premise VM for hosting the HDP Docker container.

## Tasks
1. Create the deployment directory and `.env` file.
2. Setup SSH permissions for the Gitea Runner.
3. Configure Docker to allow the Registry.

---

### Step 1: Initialize the Deployment Directory
Connect to your VM and run:

```bash
mkdir -p ~/homelab/hdp
cd ~/homelab/hdp
touch .env
```

### Step 2: Configure the `.env` file
Populate `~/homelab/hdp/.env` with your secrets (do not check this into git).

```env
# Health Data Platform - Environment Variables
SESSION_SECRET=a_very_long_random_string_here
GOOGLE_CLIENT_ID=your_id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your_secret_from_google_console
GOOGLE_CALLBACK_URL=http://your_domain_or_ip:8080/auth/google/callback
DATABASE_URL=postgres://devuser:password@192.168.1.222:5432/my_app?sslmode=disable
```

### Step 3: Handle the Insecure Registry
If your local Gitea instance isn't using HTTPS correctly on the internal network, you must allow Docker to talk to it.

On **both** the Runner machine and the Target VM, edit or create `/etc/docker/daemon.json`:

```json
{
  "insecure-registries": ["192.168.1.222:3000"]
}
```
Then restart Docker: `sudo systemctl restart docker`.

---

### Step 4: SSH Key Setup
From Phase 2, the Gitea Runner needs to reach the VM.
1. Copy the public key created earlier (`hdp_lab.pub`) into the `~/.ssh/authorized_keys` file of the user on the target VM.
2. Store the **Private Key** (`hdp_lab`) in the Gitea repo secret `SSH_PRIVATE_KEY` as mentioned in Phase 2.
