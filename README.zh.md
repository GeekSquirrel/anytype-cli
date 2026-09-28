[English](README.md) | [简体中文](README.zh.md)

> 两个语言版本必须保持同步：更新任一版本时，同步更新另一个。

# anytype-cli —— v2 API 尝鲜版（api-v2）

本 fork 用比官方锁定版本更新的 `anytype-heart` 来构建官方
[anytype-cli](https://github.com/anyproto/anytype-cli)，让自托管部署
**现在就能用上 v2 API**。CLI 本身是未改动的上游代码，只前移了内嵌的
heart 依赖，外加一项 v2 时代必需的密钥管理增强。

**退役路线**：一旦官方 anytype-cli 的 release 把 `anytype-heart` 锁到
`>= v0.51.0`，官方镜像同样带 v2——届时切回官方镜像并退役本 fork 的
workflow。

## 为什么有这个 fork

| | 官方 anytype-cli | 本 fork |
|---|---|---|
| go.mod 锁定的 heart | `v0.50.20`（仅 v1 API） | `v0.51.3`（v1 + v2 API） |
| heart 源码 | 未改动的 release | 未改动的 release |
| CLI 源码 | 未改动 | + `apikey` 的 scope/grant 参数（见下） |
| API 面 | `/v1/*` | `/v1/*` 和 `/v2/*` |

CLI 把 `anytype-heart` 作为 Go 库内嵌并复用其 `core/api` HTTP 服务（容器
端口 31012），因此 API 版本完全由编译进去的 heart 版本决定。heart
`v0.51.0` 发布了 v2 API（GO-7383）；官方 CLI 目前仍锁 `v0.50.20`。

## v2 API 新增能力（均在本构建上实测通过）

- **完整的对象面**：`GET/POST/PATCH/DELETE /v2/spaces/{space_id}/objects/...`，
  AnyBlock 文档形态（`formatVersion`/`properties`/`blocks`）、基于 etag 的
  乐观并发（`If-Match`）、幂等键与 dry-run；
- **原生内联讨论**：`POST /v2/spaces/{space_id}/objects/{object_id}/discussion`
  可为任意对象冷启动讨论，对象读取自带 `discussion` chat id——无需补丁；
- **搜索不依赖 v1**：`POST /v2/search`（跨空间）与
  `POST /v2/spaces/{space_id}/search`；配合 queries、types、properties、
  templates、members、widgets、files 端点，完整的 agent 工作流可以只跑在
  v2 上；
- **密钥自省**：`GET /v2/auth/whoami` 返回 key 的 scope 与空间授权；
- 自描述契约：`/v2/docs/openapi.json`、`/v2/schemas`。

## 本 fork 的 CLI 增强

heart v0.51.x 在 **/v2 上强制 key-scope 门禁**：没有 scope 的 key（
`Limited`，旧 CLI 的默认值）在 v2 一律被拒，`/v1` 则继续放行。本 fork 把
scope 和空间授权接入密钥命令：

```bash
# JsonAPI scope：可调用 /v1 和 /v2
anytype auth apikey create mcp --scope jsonapi

# 创建时把 key 收窄到指定空间
anytype auth apikey create mcp --scope jsonapi \
  --spaces bafyreidr2epoyudmzxf...,bafyreiedvs72vpdksf5... --perm readwrite

# 或者所有空间（动态，含未来新建的空间）
anytype auth apikey create mcp --scope jsonapi --all-spaces --perm read

# 事后原地修改 grant：key 字符串不变，客户端无需重新配置；
# heart 会清掉会话缓存，新 grant 立即生效
anytype auth apikey grant <appHash> --spaces <id,id,...> --perm read
anytype auth apikey grant <appHash> --clear     # 恢复全空间访问

# 列表现在带 SCOPE / GRANT 列
anytype auth apikey list
```

继承自 heart 的规则：grant 只存在于 `JsonAPI` key 上（`Limited` 不能带
grant；`Full` 保留给账号密钥会话，不可铸造）；`--spaces` 与
`--all-spaces` 互斥。

## 发布

`.github/workflows/release-api-v2.yml` 是手动触发（workflow_dispatch）的
构建：按上游 alpine 流程产出 linux amd64 + arm64 静态 musl 二进制，推送
多架构镜像到 ghcr。镜像 tag 为 input，默认 `api-v2`：

```
ghcr.io/geeksquirrel/anytype-cli:api-v2
```

## 在 any-sync-dockercompose 中使用

```yaml
# docker-compose.override.yml
services:
  anytype-cli:
    image: ghcr.io/geeksquirrel/anytype-cli:api-v2
  anytype-cli_bootstrap:
    image: ghcr.io/geeksquirrel/anytype-cli:api-v2
```

然后 `docker compose pull anytype-cli anytype-cli_bootstrap && docker compose up -d anytype-cli`。

> 首次用 GITHUB_TOKEN 推送的 ghcr 包默认是**私有**的。请到
> GitHub → Packages → anytype-cli → Package settings 切换为 Public，否则
> 部署机拉取前需要先 `docker login ghcr.io`。

## 旧的 mcp-enhance 补丁线

在 v2 出现之前，本 fork 曾在 heart v0.50.20 上维护 **mcp-enhance** 补丁集
（`patches/`、`release-mcp-enhance.yml`），为 v1 API 补充讨论与富文本
markdown 能力。v2 API 已原生覆盖其主要功能（对象 `discussion` 字段、讨论
冷启动、聊天消息读取）；该补丁线仍服务于 v1 时代部署，上游跟进后同样
退役。

## 上游跟进后的退役步骤

1. 关注官方 anytype-cli 的 release；当某个版本把 `anytype-heart` 锁到
   `>= v0.51.0`，官方镜像即原生带 v2；
2. 删除 `release-api-v2.yml`（以及旧的 `release-mcp-enhance.yml` /
   `patches/`），把 `release.yml` 恢复为上游版本，并将
   `docker-compose.override.yml` 指回官方镜像。
