[English](README-MCP-ENHANCED.md) | [简体中文](README-MCP-ENHANCED.zh-CN.md)

> 双语版本需同步更新，改任一语言时请同步另一份。
# 自托管补丁版（mcp-enhance）

这个 fork 在官方 anytype-cli 之上维护一套补丁集，让自托管（any-sync-dockercompose）
部署的 anytype-cli 提前获得若干面向 agent / API 的能力增强，直到官方合并对应改动。
补丁集由此得名：**mcp-enhance** —— 为 MCP / agent 工具面补全的能力。

## 它改了什么

CLI 把 `anytype-heart` 作为 Go 库引入并直接复用其 `core/api` HTTP 服务（容器 31012 端口），
所以补丁本身完全在 heart 侧，本仓库只包含让构建用上补丁 heart 的机制：

| 文件 | 作用 |
|---|---|
| `patches/anytype-heart-mcp-enhance.patch` | 补丁集，基线记录在 `patches/heart-patch-base`（当前 `v0.50.20`）。功能明细见下节 |
| `.github/workflows/release-mcp-enhance.yml` | 监控上游版本并自动构建发布补丁镜像 |
| `.github/workflows/release.yml` | 与上游一致，仅禁用 tag 触发（避免与补丁 tag 冲突）并移除 Docker Hub / Slack 步骤（fork 无对应 secrets） |

## 补丁集功能明细

### ① GO-3132：discussionId API（讨论区 Programmable）

让 agent 能像人一样读写对象的内联讨论（评论）：

- **v1 Object 模型暴露 `discussion_id`**：读取对象时能拿到其讨论区的 id
  （`discussion_id` 不在上游 OpenAPI 规范内，是本补丁的扩展字段）；
- **v1 ChatMessage 暴露 `blocks`**：聊天消息返回完整块结构，桌面端消息不再读出空文本；
- **API 消息合成 text block**：通过 API 写入的消息在服务端合成 text block，
  与桌面端发出的消息结构同构，`blocks` 成为所有消息的统一读取入口；
- **讨论冷启动端点**：新增 `POST /v1/spaces/{space_id}/objects/{object_id}/discussion`，
  agent 可自行为没有讨论区的对象创建讨论（上游仅在 UI 侧惰性创建）。

### ② Rich-markdown 往返（anymark 解析补全）

修复"导出格式吃不回去"的读写不对称——此前导出器把 Mention 标记序列化成
`anytype://` 链接、把 Mermaid 图序列化成 ` ```mermaid ` 围栏，但写入端解析器
（anymark）不认识这两种形态，导致 **agent 通过 API 写的正文永远无法还原成
真引用和真图表**（人走编辑器 RPC 不受影响）：

- **`anytype://object?objectId=…` 链接 → Mention 标记**：写入端解析时还原为
  `BlockContentTextMark_Mention`（Param = objectId）。效果：对象链接图谱正确登记
  （links / backlinks 双向可见，收集逻辑 `FillSmartIds` 只认 Mention/Object 标记）、
  UI 中渲染为对象引用卡片；普通 https 链接与非 object 的 anytype:// 链接行为不变；
- **` ```mermaid ` 围栏 → Latex 块**：解析为 `BlockContentLatex`（Processor=Mermaid），
  与上游 Notion 导入器 `handleMermaidBlock` 的构造完全一致，客户端可正常渲染图表；
  其余语言的代码围栏行为不变（仍是带 `lang` 字段的 Code 文本块）；
- **往返测试护栏**：新增 `richmd_test.go`（mention 带参/无参、mermaid、普通链接、
  普通代码围栏等回归用例），后续任何"导得出、吃不回"的块类型都应先加测试用例再修。

补丁的可追溯源码分支：[GeekSquirrel/anytype-heart `GO-3132-v0.50.20-discussion-id`](https://github.com/GeekSquirrel/anytype-heart/tree/GO-3132-v0.50.20-discussion-id)
（GO-3132 提交 + blocks 暴露 + anymark 往返补丁，基于 CLI 依赖的 heart v0.50.20）。
较新的分支 `GO-3132-expose-discussion-id`（基于更新的上游树）同步维护同一套改动，
作为将来 rebase 的前置参考。

## release-mcp-enhance.yml 工作方式

1. 每 30 分钟（cron）或手动（workflow_dispatch）解析要构建的 CLI 版本：
   手动输入 > 官方 [any-sync-dockercompose](https://github.com/anyproto/any-sync-dockercompose)
   `.env.example` 中钉住的具体版本 > **上游 anytype-cli 最新已发布的 release**（`.env` 为 `latest` 时的默认监控源，Releases API 查询，不含 prerelease）。
2. 若该版本已有 `vX.Y.Z-mcp-enhance.N` tag 则跳过（`force=true` 可强制追加新编号）。
   注：命名从 `vX.Y.Z-discussion.N` 迁移而来，旧 tag 不参与计数，新编号从 .1 重新起算。
3. 从该 CLI 版本的 `go.mod` 解析其依赖的 heart 版本（tag 或伪版本自动转 commit SHA），
   检出**上游** anytype-heart 对应版本并应用补丁集
   （已包含于上游则自动跳过补丁；上下文漂移时用 `git apply -3` 兜底；补丁失配或编译失败都会
   硬失败、不发布镜像，等待手动 rebase `patches/` 后重跑）。
4. `go mod replace` 指向补丁 heart，按上游同款 alpine/musl 流程静态编译 linux amd64 + arm64，
   推送镜像 `ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-mcp-enhance.N` 和移动 tag `mcp-enhance`，
   并创建同名 GitHub Release（附 linux 二进制）。

## 在 any-sync-dockercompose 中使用

在部署目录建 `docker-compose.override.yml`：

```yaml
services:
  anytype-cli:
    image: ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-mcp-enhance.N
  anytype-cli_bootstrap:
    image: ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-mcp-enhance.N
```

然后 `docker compose pull anytype-cli anytype-cli_bootstrap && docker compose up -d anytype-cli`。

> 注意：GITHUB_TOKEN 推送的首个 ghcr 包默认是**私有**的。到 GitHub → Packages →
> anytype-cli → Package settings 里改成 Public，否则部署机拉取需要先 `docker login ghcr.io`。

> 从旧 `discussion` 镜像 tag 迁移：新镜像的移动 tag 是 `mcp-enhance`，
> compose.override 里的镜像引用需同步更新。

## 上游合并后如何退役

补丁集对应的改动被官方合并并发布后：

1. workflow 检测到补丁"已包含于上游"，之后发布的 `mcp-enhance.N` 就是纯净上游构建（行为不变，可继续当镜像跟随器用）；
2. 彻底清理：删除 `patches/`、`release-mcp-enhance.yml`，恢复 `release.yml` 为上游版本，
   删除 heart fork 的补丁分支，compose.override 改回官方镜像。
