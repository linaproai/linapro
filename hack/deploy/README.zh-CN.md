# LinaPro 演示部署

该目录包含用于启动一次性 LinaPro 演示环境的`Docker Compose`资源。

## 文件说明

| 文件 | 用途 |
| --- | --- |
| `docker-compose.yaml` | 启动 PostgreSQL 与 LinaPro nightly 镜像，用于本地演示。 |
| `config.yaml` | 提供挂载到 LinaPro 容器内的运行时配置。 |

## 前置条件

- 已安装支持`docker compose`插件的`Docker`。
- 能够拉取`postgres:14-alpine`与`ghcr.io/linaproai/linapro:nightly`镜像，或本地已经存在这些镜像。
- 主机端口`9120`可用，或提供自定义`LINAPRO_HTTP_PORT`值。

## 启动演示环境

在仓库根目录执行：

```bash
docker compose -f hack/deploy/docker-compose.yaml up
```

等待`linapro`服务启动完成后，访问`http://localhost:9120`。

如需后台运行：

```bash
docker compose -f hack/deploy/docker-compose.yaml up -d
```

## 运行行为

该`compose`文件用于可重复的演示环境，不适合作为生产部署。

- PostgreSQL 数据存放在`tmpfs`中。
- LinaPro 的`/app/data`运行时数据存放在`tmpfs`中。
- LinaPro 容器每次启动都会执行`./lina init --confirm=init --rebuild=true`。
- 每次启动都会通过`./lina mock --confirm=mock`加载 mock 数据。
- 容器删除或重建后，数据会丢失。

除非明确调整存储模型，否则不要把该部署方式用于需要持久化数据的环境。

## 配置说明

`config.yaml`会以只读方式挂载到 LinaPro 容器内的`/run/linapro/config.yaml`。
容器通过`GF_GCFG_PATH=/run/linapro`让 LinaPro 在启动时读取该配置文件。

常用配置：

| 配置项 | 默认值 | 说明 |
| --- | --- | --- |
| `server.address` | `:9120` | 容器内部监听地址。 |
| `database.default.link` | compose 网络中的 PostgreSQL 服务 | 与`postgres`服务凭据保持一致。 |
| `jwt.secret` | 演示专用密钥 | 一次性演示以外的环境必须替换。 |
| `i18n.default` | `zh-CN` | 演示环境默认语言。 |
| `cluster.enabled` | `false` | 该演示只启动一个 LinaPro 容器。 |
| `plugin.autoEnable` | 内置演示插件 | 启用宿主与源码插件演示能力，并加载 mock 数据。 |

## 环境变量覆盖

| 变量 | 默认值 | 用途 |
| --- | --- | --- |
| `LINAPRO_HTTP_PORT` | `9120` | 映射到 LinaPro 容器`9120`端口的主机端口。 |
| `LINAPRO_SERVER_ADDRESS` | （未设置） | 设置后覆盖进程监听地址（`server.address`），例如 `:18080` 或 `0.0.0.0:9120`。不替代 Docker 宿主机端口映射。 |
| `LINAPRO_IMAGE` | `ghcr.io/linaproai/linapro:nightly` | 要运行的 LinaPro 镜像。 |
| `LINAPRO_POSTGRES_IMAGE` | `postgres:14-alpine` | 要运行的 PostgreSQL 兼容镜像。 |

示例：

```bash
LINAPRO_HTTP_PORT=18080 docker compose -f hack/deploy/docker-compose.yaml up
```

```bash
LINAPRO_IMAGE=ghcr.io/linaproai/linapro:nightly docker compose -f hack/deploy/docker-compose.yaml up
```

## 常用命令

查看服务状态：

```bash
docker compose -f hack/deploy/docker-compose.yaml ps
```

查看日志：

```bash
docker compose -f hack/deploy/docker-compose.yaml logs -f linapro
```

停止演示环境：

```bash
docker compose -f hack/deploy/docker-compose.yaml down
```

重新创建全新的演示环境：

```bash
docker compose -f hack/deploy/docker-compose.yaml down
docker compose -f hack/deploy/docker-compose.yaml up
```
