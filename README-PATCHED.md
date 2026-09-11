# 自托管补丁版（discussion-id）

这个 fork 在官方 anytype-cli 之上维护一个临时补丁，让自托管（any-sync-dockercompose）
部署的 anytype-cli 提前支持 [GO-3132 discussionId API](https://github.com/GeekSquirrel/anytype-heart/tree/GO-3132-expose-discussion-id)，
直到官方合并该改动。

## 它改了什么

CLI 把 `anytype-heart` 作为 Go 库引入并直接复用其 `core/api` HTTP 服务（容器 31012 端口），
所以补丁本身完全在 heart 侧，本仓库只包含让构建用上补丁 heart 的机制：

| 文件 | 作用 |
|---|---|
| `patches/anytype-heart-discussion-id.patch` | 补丁，基线记录在 `patches/heart-patch-base`（当前 `v0.50.20`）：① GO-3132 discussionId API；② v1 ChatMessage 暴露 blocks（桌面端消息不再读出空文本）；③ v1 写入侧合成 text block（API 发的消息与桌面端同构，`blocks` 成为所有消息的统一读取入口） |
| `.github/workflows/release-discussion.yml` | 监控上游版本并自动构建发布补丁镜像 |
| `.github/workflows/release.yml` | 与上游一致，仅禁用 tag 触发（避免与补丁 tag 冲突）并移除 Docker Hub / Slack 步骤（fork 无对应 secrets） |

补丁的可追溯源码分支：[GeekSquirrel/anytype-heart `GO-3132-v0.50.20-discussion-id`](https://github.com/GeekSquirrel/anytype-heart/tree/GO-3132-v0.50.20-discussion-id)
（GO-3132 提交 + blocks 暴露，移植到 CLI 依赖的 heart v0.50.20 上）。

## release-discussion.yml 工作方式

1. 每 30 分钟（cron）或手动（workflow_dispatch）解析要构建的 CLI 版本：
   手动输入 > 官方 [any-sync-dockercompose](https://github.com/anyproto/any-sync-dockercompose)
   `.env.example` 中钉住的具体版本 > **上游 anytype-cli 最新已发布的 release**（`.env` 为 `latest` 时的默认监控源，Releases API 查询，不含 prerelease）。
2. 若该版本已有 `vX.Y.Z-discussion.N` tag 则跳过（`force=true` 可强制追加新编号）。
3. 从该 CLI 版本的 `go.mod` 解析其依赖的 heart 版本（tag 或伪版本自动转 commit SHA），
   检出**上游** anytype-heart 对应版本并应用补丁
   （已包含于上游则自动跳过补丁；上下文漂移时用 `git apply -3` 兜底；补丁失配或编译失败都会
   硬失败、不发布镜像，等待手动 rebase `patches/` 后重跑）。
4. `go mod replace` 指向补丁 heart，按上游同款 alpine/musl 流程静态编译 linux amd64 + arm64，
   推送镜像 `ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-discussion.N` 和移动 tag `discussion`，
   并创建同名 GitHub Release（附 linux 二进制）。

## 在 any-sync-dockercompose 中使用

在部署目录建 `docker-compose.override.yml`：

```yaml
services:
  anytype-cli:
    image: ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-discussion.N
  anytype-cli_bootstrap:
    image: ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-discussion.N
```

然后 `docker compose pull anytype-cli anytype-cli_bootstrap && docker compose up -d anytype-cli`。

> 注意：GITHUB_TOKEN 推送的首个 ghcr 包默认是**私有**的。到 GitHub → Packages →
> anytype-cli → Package settings 里改成 Public，否则部署机拉取需要先 `docker login ghcr.io`。

## 上游合并后如何退役

GO-3132 被官方合并并发布后：

1. workflow 检测到补丁"已包含于上游"，之后发布的 `discussion.N` 就是纯净上游构建（行为不变，可继续当镜像跟随器用）；
2. 彻底清理：删除 `patches/`、`release-discussion.yml`，恢复 `release.yml` 为上游版本，
   删除 `v0.50.20-discussion-id` heart 分支，compose.override 改回官方镜像。
