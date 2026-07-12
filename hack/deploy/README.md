# LinaPro Demo Deployment

This directory contains the `Docker Compose` assets for starting a disposable LinaPro demo environment.

## Files

| File | Purpose |
| --- | --- |
| `docker-compose.yaml` | Starts PostgreSQL and the LinaPro nightly image for a local demo. |
| `config.yaml` | Provides the runtime configuration mounted into the LinaPro container. |

## Prerequisites

- `Docker` with the `docker compose` plugin.
- Network access to pull `postgres:14-alpine` and `ghcr.io/linaproai/linapro:nightly`, unless the images already exist locally.
- Host port `9120` available, or a custom `LINAPRO_HTTP_PORT` value.

## Start the Demo

Run the command from the repository root:

```bash
docker compose -f hack/deploy/docker-compose.yaml up
```

Open `http://localhost:9120` after the `linapro` service finishes startup.

To run the demo in the background:

```bash
docker compose -f hack/deploy/docker-compose.yaml up -d
```

## Runtime Behavior

The compose file is intended for repeatable demos, not production deployment.

- PostgreSQL data is stored in `tmpfs`.
- LinaPro runtime data under `/app/data` is stored in `tmpfs`.
- The LinaPro container runs `./lina init --confirm=init --rebuild=true` on every start.
- Mock data is loaded on every start with `./lina mock --confirm=mock`.
- Data disappears after the containers are removed or recreated.

Do not use this deployment as a persistent environment unless the storage model is intentionally changed.

## Configuration

`config.yaml` is mounted read-only into the LinaPro container at `/run/linapro/config.yaml`.
The container uses `GF_GCFG_PATH=/run/linapro` so LinaPro reads that mounted configuration file at startup.

Common settings:

| Setting | Default | Notes |
| --- | --- | --- |
| `server.address` | `:9120` | Internal container address. |
| `database.default.link` | PostgreSQL service in the compose network | Matches the `postgres` service credentials. |
| `jwt.secret` | Demo-only secret | Replace it outside disposable demos. |
| `i18n.default` | `zh-CN` | Default demo language. |
| `cluster.enabled` | `false` | This demo starts one LinaPro container. |
| `plugin.autoEnable` | Built-in demo plugins | Enables host and source-plugin demo surfaces with mock data. |

## Environment Overrides

| Variable | Default | Purpose |
| --- | --- | --- |
| `LINAPRO_HTTP_PORT` | `9120` | Host port mapped to the LinaPro container port `9120`. |
| `LINAPRO_SERVER_ADDRESS` | (unset) | When set, overrides process listen address (`server.address`), e.g. `:18080` or `0.0.0.0:9120`. Does not replace Docker host port mapping. |
| `LINAPRO_IMAGE` | `ghcr.io/linaproai/linapro:nightly` | LinaPro image to run. |
| `LINAPRO_POSTGRES_IMAGE` | `postgres:14-alpine` | PostgreSQL-compatible image to run. |

Example:

```bash
LINAPRO_HTTP_PORT=18080 docker compose -f hack/deploy/docker-compose.yaml up
```

```bash
LINAPRO_IMAGE=ghcr.io/linaproai/linapro:nightly docker compose -f hack/deploy/docker-compose.yaml up
```

## Useful Commands

Check service status:

```bash
docker compose -f hack/deploy/docker-compose.yaml ps
```

Follow logs:

```bash
docker compose -f hack/deploy/docker-compose.yaml logs -f linapro
```

Stop the demo:

```bash
docker compose -f hack/deploy/docker-compose.yaml down
```

Recreate a fresh demo:

```bash
docker compose -f hack/deploy/docker-compose.yaml down
docker compose -f hack/deploy/docker-compose.yaml up
```
