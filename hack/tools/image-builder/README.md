# Image Builder Tool

`image-builder` builds the production LinaPro Docker image from the standard `make build` output and the repository root configuration file `hack/config.yaml`. It is the Docker execution layer behind `make image`.

## Usage

Preferred repository entry point:

```bash
make image
make image tag=v0.6.0
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1
make image os=linux arch=amd64
make image platform=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v0.6.0 push=1
```

Direct tool invocation:

```bash
make build
go run ./hack/tools/image-builder --tag=v0.6.0
go run ./hack/tools/image-builder --tag=v0.6.0 --registry=ghcr.io/linaproai --push=1
```

Cross-platform host binaries can be built with `make build os=linux arch=arm64`. Multi-platform image publishing uses `Docker buildx`; for example:

```bash
make image platform=linux/amd64,linux/arm64 registry=ghcr.io/linaproai tag=v0.6.0 push=1
```

## Configuration

Build defaults are read from `hack/config.yaml` under the `build` section.

| Field | Description |
| --- | --- |
| `os` / `arch` / `platform` | Target host binary and Docker image platform list. `auto` follows the local Go architecture for `arch` and resolves `platform` to `<os>/<arch>`. Use comma-separated `platform` values for multi-platform builds. |
| `cgoEnabled` | Whether `make build` enables CGO for the host binary. |
| `outputDir` / `binaryName` | Repository-relative standard `make build` artifact location. |

Image metadata defaults are read from `hack/config.yaml` under the `image` section.

| Field | Description |
| --- | --- |
| `name` | Docker image repository name, without registry prefix. |
| `tag` | Optional default tag. Empty means derive from `git describe`. |
| `registry` | Optional remote registry prefix, such as `ghcr.io/linaproai`. |
| `push` | Default push behavior. |
| `baseImage` | Runtime base image passed to the Dockerfile. |
| `dockerfile` | Repository-relative Dockerfile path. Defaults to `hack/docker/Dockerfile`. |

Command-line flags override the config file for one invocation. `LINAPRO_IMAGE_REGISTRY` can also provide the registry prefix when neither the config nor `registry=...` is set.

Repository structure paths such as `apps/lina-core`, `apps/lina-vben`, `apps/lina-plugins`, embedded public assets, packed manifest assets, and the `build-wasm` tool path are project conventions and are intentionally not exposed in `hack/config.yaml`.

## Output

- `make build` copies frontend production assets into the host embed workspace.
- `make build` prepares host manifest assets without embedding local `config.yaml`.
- `make build` writes dynamic plugin `Wasm` artifacts into the configured build output directory.
- `make build` compiles the host binary for the configured target platform.
- `make build platform=linux/amd64,linux/arm64` writes host binaries into `temp/output/linux_amd64/lina` and `temp/output/linux_arm64/lina`.
- `make image` stages the standard host binary into the Docker build context instead of rebuilding it.
- Single-platform Docker builds use `docker build`; multi-platform Docker builds use `docker buildx build --push`.
- Docker builds `<registry-prefix>/<name>:<tag>` and only pushes when `push=true`. Multi-platform builds require `push=true` so the remote manifest is published.
