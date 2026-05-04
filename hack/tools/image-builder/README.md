# Image Builder Tool

`image-builder` builds the production LinaPro Docker image from the repository root configuration file `hack/config.yaml`. It is the cross-platform execution layer behind `make image`.

## Usage

Preferred repository entry point:

```bash
make image
make image tag=v0.6.0
make image tag=v0.6.0 registry=ghcr.io/linaproai push=1
```

Direct tool invocation:

```bash
go run ./hack/tools/image-builder --tag=v0.6.0
go run ./hack/tools/image-builder --tag=v0.6.0 --registry=ghcr.io/linaproai --push=1
```

## Configuration

Defaults are read from `hack/config.yaml` under the `image` section.

| Field | Description |
| --- | --- |
| `name` | Docker image repository name, without registry prefix. |
| `tag` | Optional default tag. Empty means derive from `git describe`. |
| `registry` | Optional remote registry prefix, such as `ghcr.io/linaproai`. |
| `push` | Default push behavior. |
| `baseImage` | Runtime base image passed to the Dockerfile. |
| `os` / `arch` / `platform` | Target binary and Docker image platform. `auto` follows the local Go architecture. |
| `dockerfile` | Repository-relative Dockerfile path. Defaults to `hack/docker/Dockerfile`. |
| `outputDir` / `binaryName` | Repository-relative image build artifact location. |

Command-line flags override the config file for one invocation. `LINAPRO_IMAGE_REGISTRY` can also provide the registry prefix when neither the config nor `registry=...` is set.

Repository structure paths such as `apps/lina-core`, `apps/lina-vben`, `apps/lina-plugins`, embedded public assets, packed manifest assets, and the `build-wasm` tool path are project conventions and are intentionally not exposed in `hack/config.yaml`.

## Output

- Frontend production assets are copied into the host embed workspace.
- Host manifest assets are prepared without embedding local `config.yaml`.
- Dynamic plugin `Wasm` artifacts are written into the configured output directory.
- The host binary is compiled with `CGO_ENABLED=0` for the configured target platform.
- Docker builds `<registry-prefix>/<name>:<tag>` and only pushes when `push=true`.
