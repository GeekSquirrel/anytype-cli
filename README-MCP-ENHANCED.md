[English](README-MCP-ENHANCED.md) | [简体中文](README-MCP-ENHANCED.zh-CN.md)

> Both language versions must be kept in sync: when updating either one, update the other.

# Self-hosted patched build (mcp-enhance)

This fork maintains a patch set on top of the official anytype-cli so that
self-hosted (any-sync-dockercompose) deployments get several agent/API-facing
capability enhancements ahead of upstream. The patch set is named
**mcp-enhance**: capabilities added for the MCP / agent tool surface.

## What it changes

The CLI embeds `anytype-heart` as a Go library and reuses its `core/api` HTTP
service (container port 31012), so all patches live on the heart side; this
repo only carries the mechanism that makes the build use the patched heart:

| File | Purpose |
|---|---|
| `patches/anytype-heart-mcp-enhance.patch` | The patch set; its base is recorded in `patches/heart-patch-base` (currently `v0.50.20`). See details below |
| `.github/workflows/release-mcp-enhance.yml` | Watches upstream versions and builds/publishes the patched image automatically |
| `.github/workflows/release.yml` | Same as upstream, except the tag trigger is disabled (avoids clashing with patch tags) and Docker Hub / Slack steps are removed (the fork lacks those secrets) |

## Patch set features in detail

### 1. GO-3132: discussionId API (programmable discussions)

Lets agents read and write object inline discussions (comments) like humans do:

- **v1 Object models expose `discussion_id`**: reading an object returns the id
  of its inline discussion (`discussion_id` is not part of the upstream OpenAPI
  spec — it is an extension field added by this patch);
- **v1 ChatMessage exposes `blocks`**: chat messages return their full block
  structure, so desktop-originated messages no longer read back as empty text;
- **API messages get synthesized text blocks**: messages created through the
  API get a server-side synthesized text block, making them structurally
  identical to desktop-originated ones — `blocks` becomes the single uniform
  read path for all messages;
- **Discussion cold-start endpoint**: new endpoint
  `POST /v1/spaces/{space_id}/objects/{object_id}/discussion`, so agents can
  create a discussion for objects that don't have one yet (upstream only
  creates discussions lazily from the UI).

### 2. Rich-markdown round-trip (anymark parsing fix)

Fixes the read/write asymmetry of "exported formats that cannot be parsed
back": the exporters serialize Mention marks as `anytype://` links and Mermaid
diagrams as ` ```mermaid ` fences, but the write-side parser (anymark) did not
understand either form — so **rich content written by agents through the API
could never become real mentions or real diagrams** (humans are unaffected
because the editor talks RPC directly):

- **`anytype://object?objectId=…` links → Mention marks**: parsed back into
  `BlockContentTextMark_Mention` (Param = objectId) on write. Effect: the
  object link graph registers correctly (links / backlinks both visible —
  the collector `FillSmartIds` only counts Mention/Object marks) and the UI
  renders an object mention card; plain https links and non-object
  anytype:// links behave exactly as before;
- **` ```mermaid ` fences → Latex blocks**: parsed into `BlockContentLatex`
  (Processor=Mermaid), matching the upstream Notion importer's
  `handleMermaidBlock` construction exactly, so clients render diagrams;
  code fences in all other languages are unchanged (still Code text blocks
  with a `lang` field);
- **Round-trip test guardrails**: new `richmd_test.go` (mention with/without
  spaceId, mermaid, plain links, plain code fences as regression cases). Any
  future "exportable but not parseable" block type should get a test case
  first, then a fix.

Traceable source branch for the patches:
[GeekSquirrel/anytype-heart `GO-3132-v0.50.20-discussion-id`](https://github.com/GeekSquirrel/anytype-heart/tree/GO-3132-v0.50.20-discussion-id)
(GO-3132 commits + blocks exposure + the anymark round-trip fix, based on
heart v0.50.20 which is what the CLI depends on). The newer branch
`GO-3132-expose-discussion-id` (based on a newer upstream tree) carries the
same changes and serves as a reference for future rebases.

## How release-mcp-enhance.yml works

1. Every 30 minutes (cron) or on manual dispatch (workflow_dispatch), resolve
   the CLI version to build: manual input > a concrete version pinned in the
   official [any-sync-dockercompose](https://github.com/anyproto/any-sync-dockercompose)
   `.env.example` > **the latest published release of upstream anytype-cli**
   (the default watch source when `.env` says `latest`; via the Releases API,
   prereleases excluded).
2. If a `vX.Y.Z-mcp-enhance.N` tag already exists for that version, skip
   (`force=true` forces a new number). Note: the naming was migrated from
   `vX.Y.Z-discussion.N`; old tags are not counted and numbering restarts at .1.
3. Read the heart version required by that CLI release from its `go.mod`
   (tags and pseudo-versions are auto-converted to commit SHAs), check out
   **upstream** anytype-heart at that version and apply the patch set
   (auto-skipped if already contained upstream; `git apply -3` as a fallback
   on context drift; patch mismatch or build failure hard-fails without
   publishing an image, waiting for a manual rebase of `patches/`).
4. `go mod replace` points at the patched heart, then a static linux amd64 +
   arm64 musl build follows the upstream alpine flow; pushes
   `ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-mcp-enhance.N` plus the moving tag
   `mcp-enhance`, and creates a GitHub Release of the same name (with linux
   binaries attached).

## Using it in any-sync-dockercompose

Create a `docker-compose.override.yml` in the deployment directory:

```yaml
services:
  anytype-cli:
    image: ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-mcp-enhance.N
  anytype-cli_bootstrap:
    image: ghcr.io/geeksquirrel/anytype-cli:vX.Y.Z-mcp-enhance.N
```

Then `docker compose pull anytype-cli anytype-cli_bootstrap && docker compose up -d anytype-cli`.

> Note: the first ghcr package pushed with GITHUB_TOKEN is **private** by
> default. Go to GitHub → Packages → anytype-cli → Package settings and switch
> it to Public, or pulling from the deployment host requires `docker login ghcr.io` first.

> Migrating from the old `discussion` image tag: the moving tag is now
> `mcp-enhance`; update the image references in your compose.override accordingly.

## Retiring once upstream merges

Once the patched changes are merged and released upstream:

1. The workflow detects the patch is "already in upstream", so subsequent
   `mcp-enhance.N` builds become plain upstream builds (behavior unchanged —
   they can keep serving as an image follower);
2. Full cleanup: delete `patches/` and `release-mcp-enhance.yml`, restore
   `release.yml` to the upstream version, delete the heart fork's patch
   branches, and point compose.override back at the official image.
