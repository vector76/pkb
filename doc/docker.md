# Running PKB in Docker

Docker Compose runs PKB and Raymond as two containers sharing your
knowledge base from the host.  The KB is its own git repo; credentials,
workflow files, and Docker config live alongside it but outside its
git history.

## Directory layout

Set up a directory with this structure:

```
pkb-docker/
  .claude-auth/         Claude credentials (outside KB git)
  Dockerfile            Container build definition
  docker-compose.yml    Service orchestration
  .dockerignore         Build context filter
  ray_wf/               Raymond workflow files (outside KB git)
    START.md
    WAIT.sh
    REPLY_1.sh
    ...
  kb/                   Knowledge base (its own git repo)
    .git/
    .raymond/config.toml
    wiki/
    conversations/
    attachments/
    queue/
    ...
```

Inside the container, `ray_wf/` is mounted at `/ray_wf` — completely
separate from the KB at `/kb`.

## Prerequisites

- Docker and Docker Compose (v2)
- An Anthropic account (for Claude Code authentication)

## Quick start

```bash
# 1. Create the directory structure
mkdir pkb-docker && cd pkb-docker
mkdir .claude-auth

# 2. Copy Dockerfile, docker-compose.yml, and .dockerignore into pkb-docker/
# 3. Copy ray_wf/ into pkb-docker/ray_wf/
# 4. Clone or move your knowledge base into pkb-docker/kb/

# 5. Build the image (compiles pkb + raymond, installs Claude Code CLI)
docker compose build

# 6. Authenticate Claude Code (one-time, interactive)
docker compose run --rm raymond claude
# Follow the OAuth prompts.  Credentials are saved to .claude-auth/.

# 7. Start both services
docker compose up
```

Open <http://localhost:4242> to access PKB.

## Volumes

| Host path | Container path | Mode | Purpose |
|-----------|---------------|------|---------|
| `./kb` | `/kb` | read-write | Knowledge base git repo |
| `./ray_wf` | `/ray_wf` | read-only | Raymond workflow definitions |
| `./.claude-auth` | `/home/pkb/.claude` | read-write | Claude Code credentials |

## Daily usage

```bash
# Start in the background
docker compose up -d

# Watch Raymond output (like running it in a terminal)
docker compose logs -f raymond

# Restart Raymond if it crashes or pauses
docker compose restart raymond

# Stop everything
docker compose down
```

## Re-authenticating Claude

If your Claude session expires:

```bash
docker compose run --rm raymond claude
```

## Raymond config

Raymond reads `.raymond/config.toml` from the KB root.  Since `kb/` is
on the host, edit that file directly — changes take effect on the next
Raymond restart.

## Troubleshooting

### Raymond shows as offline in PKB UI

```bash
docker compose ps
docker compose logs raymond
```

If it exited, restart it with `docker compose restart raymond`.

### Permission errors on mounted volumes

The container runs as the `pkb` user (non-root).  Make sure the host
directories are readable and writable by the container's UID:

```bash
# Check the UID inside the container
docker compose run --rm raymond id

# Fix ownership on the host if needed
sudo chown -R <uid>:<gid> kb/ .claude-auth/
```

### Build fails on Go version

The Dockerfile uses `golang:1.25-alpine`.  If that image isn't
available yet, change it to the latest available (e.g. `golang:1.24-alpine`).
