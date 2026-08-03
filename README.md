# ssh-wrapper

A minimal SSH wrapper designed for AI agents running in "yolo mode" — where the agent has full autonomy including ssh.

## Why This Exists

When an AI agent has full shell access, it can accidentally (or intentionally) access SSH keys and connect to arbitrary hosts. This wrapper prevents that by:

- **Blocking unauthorized hosts** — Only hosts in the allowlist can be reached
- **Hiding SSH keys from the agent** — The agent never sees the private key; the wrapper invokes SSH with the key directly
- **Tamper detection** — Any modification to config, key, or SSH binary causes immediate exit

## How It Works

The wrapper replaces `/usr/bin/ssh` inside the container. Every SSH call made by the agent (including those triggered by `git clone`, `git push`, etc.) passes through the wrapper first.

On each invocation it:

1. Verifies that the config file, private key, and the real `ssh` binary are all owned by root with strict permissions — any tampering causes an immediate exit
2. Logs the full command with a timestamp to the configured log path
3. Checks the target host against the allowlist in `config.yaml` — if the host is not listed, the call is denied with `Access Denied` and exits non-zero
4. Clears the environment and re-invokes the real SSH binary using the managed private key, so the agent never has direct access to the key itself

The binary runs with the `setuid` bit set, allowing it to read root-owned secrets even when invoked by an unprivileged user (uid 1000).

## Security Model

| What is protected | How                                                          |
| ----------------- | ------------------------------------------------------------ |
| Private key       | Owned by root, mode 0400, never exposed to the agent process |
| Config file       | Same — root-owned, tamper detection on startup               |
| SSH binary        | Integrity check on startup                                   |
| Unknown hosts     | Denied before any network connection is made                 |
| Environment       | Cleared before exec — no agent-injected env vars reach ssh   |

The agent can only reach hosts explicitly listed in `config.yaml`. Everything else is blocked and logged.

## Configuration

Mount two files into the container as root-owned secrets:

**`/etc/ssh.config.yaml`** — the config of the ssh-wrapper, mode 0400, owned by root.

```yaml
logpath: /var/log/ssh-wrapper/ssh-wrapper.log

allowed:
  - host: ghhy                      # Host alias (used in git clone git@ghhy:...)
    hostname: github.com            # Real hostname sent to SSH
    key_path: /etc/keys/ghhy_key    # Custom key for this host (mode 0400, owned by root)
    path_prefix:
      - elmhuangyu/

  - host: github.com                # Fallback to default key (/etc/keys/key)
    path_prefix:
      - elmhuangyu/
```

**`/etc/keys/`** — root-owned directory (`0700`) containing keys, where `/etc/keys/key` is the default private key (`0400`).

## Docker & Asgard Usage

The recommended approach to configure `ssh-wrapper` (for Asgard or other custom agent images) is to write a **custom Dockerfile** extending `asgard` (or `asgard-base-devtool`).

In this Dockerfile:
1. Copy the configuration file to `/etc/ssh.config.yaml` and set its ownership to `root:root` and permission to `0400`.
2. Copy SSH private keys to `/etc/keys/` and set ownership to `root:root` with directory permission `0700` and key file permission `0400`.
3. Create the log directory (e.g. `/var/log/ssh-wrapper`) if specified in `ssh.config.yaml`.
4. Switch to the unprivileged user (`USER user`) before running Asgard.

### Example Custom Dockerfile

```dockerfile
FROM ghcr.io/agentdrasil/asgard:latest

# 1. Copy config and set strict root-only permissions (0400)
COPY config/ssh.config.yaml /etc/ssh.config.yaml
RUN chown root:root /etc/ssh.config.yaml && \
    chmod 0400 /etc/ssh.config.yaml

# 2. Copy SSH keys into /etc/keys/ with strict permissions
COPY keys/ /etc/keys/
RUN chown -R root:root /etc/keys && \
    chmod 0700 /etc/keys && \
    chmod 0400 /etc/keys/*

# 3. Prepare log directory
RUN mkdir -p /var/log/ssh-wrapper && \
    chmod 777 /var/log/ssh-wrapper

# 4. Drop privileges to non-root user
USER user

# 5. Start Asgard
CMD ["asgard"]
```

### Example `ssh.config.yaml`

```yaml
logpath: /var/log/ssh-wrapper/ssh-wrapper.log

allowed:
  - host: ghhy                      # Host alias (used in git clone git@ghhy:...)
    hostname: github.com            # Real hostname sent to SSH
    key_path: /etc/keys/ghhy_key    # Custom key for this host (mode 0400, owned by root)
    path_prefix:
      - my-org/

  - host: github.com                # Host matching github.com (uses default key /etc/keys/key)
    path_prefix:
      - my-org/
```

## E2E Tests

Tests run entirely locally via Docker Compose — no secrets are stored anywhere. A fresh SSH key pair is generated on every run.

```bash
uvx pytest -v --log-cli-level=INFO -s
```

The test suite spins up two containers: `test-app` (the wrapper image) and `git-server` (a local SSH git server). It verifies that:

- `git clone`, `git push`, and `git pull` succeed against the allowed host
- SSH to a non-allowlisted host is blocked with `Access Denied`
- All activity is written to the log file

Tests also run in GitHub Actions on every push and pull request, with no secrets required.

## File Structure

```
.
├── main.go                  # wrapper entrypoint
├── lib/
│   ├── command/             # allowlist enforcement
│   ├── config/              # config parsing
│   └── files/               # security verification
├── Dockerfile
├── test-compose.yaml        # e2e test environment
├── docker-entrypoint.sh     # sets permissions, drops to uid 1000
└── e2e/
    └── test_e2e.py          # test runner
```

## License

Apache 2.0
