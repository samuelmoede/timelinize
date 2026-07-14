# Self-hosted runner for the Portainer redeploy step

`build-docker.yml`'s `redeploy` job needs `runs-on: [self-hosted, synology]`
instead of a GitHub-hosted runner because it talks to Portainer on a private
LAN IP (`192.168.178.20:9000`) that GitHub's cloud runners cannot reach, and
Portainer Community Edition has no stack/container webhooks (Business
Edition only) to trigger a redeploy from outside that network.

## Setup

Runs as a Portainer stack alongside the Timelinize stack itself, using the
[`myoung34/github-runner`](https://github.com/myoung34/docker-github-actions-runner)
image:

```yaml
services:
  github-runner:
    image: myoung34/github-runner:latest
    container_name: github-runner-timelinize
    restart: unless-stopped
    environment:
      REPO_URL: https://github.com/samuelmoede/timelinize
      ACCESS_TOKEN: "<fine-grained GitHub PAT, Administration: read/write on this repo>"
      RUNNER_NAME: synology-runner
      RUNNER_WORKDIR: /tmp/runner/work
      LABELS: self-hosted,synology
      EPHEMERAL: "false"
    volumes:
      - github-runner-work:/tmp/runner/work

volumes:
  github-runner-work:
```

`ACCESS_TOKEN` (a PAT) lets the container mint its own short-lived runner
registration token on every start, so it survives restarts without manual
token refresh. Verify registration at
`https://github.com/samuelmoede/timelinize/settings/actions/runners`.

The `build-and-push` job stays on `ubuntu-latest` (GitHub-hosted) — compiling
libvips and Go is heavy and GitHub's cloud minutes are free; only the final
Portainer API call needs to run inside the LAN.
