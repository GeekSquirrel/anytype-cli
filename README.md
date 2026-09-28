[English](README.md) | [简体中文](README.zh.md)

> Both language versions must be kept in sync: when updating either one, update the other.

# anytype-cli — v2 API preview build (api-v2)

This fork builds the official [anytype-cli](https://github.com/anyproto/anytype-cli)
against a newer `anytype-heart` than upstream pins, so self-hosted deployments
get the **v2 API** today. The CLI itself is unmodified upstream code; only the
embedded heart dependency moves forward, plus a small key-management addition
that v2 makes necessary.

**Retirement story:** once upstream anytype-cli pins `anytype-heart >= v0.51.0`
in its own releases, the official image serves v2 too — switch back to it and
retire this fork's workflows.

## Why this fork exists

| | upstream anytype-cli | this fork |
|---|---|---|
| heart pinned in go.mod | `v0.50.20` (v1 API only) | `v0.51.3` (v1 + v2 API) |
| heart source | unmodified release | unmodified release |
| CLI source | unmodified | + `apikey` scope/grant flags (see below) |
| API surface | `/v1/*` | `/v1/*` and `/v2/*` |

The CLI embeds `anytype-heart` as a Go library and reuses its `core/api` HTTP
service (container port 31012), so the API version is decided entirely by the
heart version compiled in. heart `v0.51.0` shipped the v2 API (GO-7383); the
official CLI still pins `v0.50.20`.

## What the v2 API adds (all verified working on this build)

- **Full object surface**: `GET/POST/PATCH/DELETE /v2/spaces/{space_id}/objects/...`
  with the AnyBlock document form (`formatVersion`/`properties`/`blocks`),
  etag-based optimistic concurrency (`If-Match`), idempotency keys and dry-run;
- **Inline discussions natively**: `POST /v2/spaces/{space_id}/objects/{object_id}/discussion`
  cold-starts a discussion on any object, and object reads expose the
  `discussion` chat id — no patch needed;
- **Search without v1**: `POST /v2/search` (cross-space) and
  `POST /v2/spaces/{space_id}/search`; together with queries, types,
  properties, templates, members, widgets and file endpoints, a complete
  agent loop runs on v2 alone;
- **Key introspection**: `GET /v2/auth/whoami` reports the key's scope and
  space grant;
- Self-describing contracts: `/v2/docs/openapi.json`, `/v2/schemas`.

## CLI additions in this fork

heart v0.51.x enforces a **key-scope gate on /v2**: keys minted without a
scope (`Limited`, the old CLI default) are refused there, while `/v1` keeps
serving them. This fork threads scope and space grants through the key
commands:

```bash
# JsonAPI scope: may call /v1 and /v2
anytype auth apikey create mcp --scope jsonapi

# Narrow the key to specific spaces at creation time
anytype auth apikey create mcp --scope jsonapi \
  --spaces bafyreidr2epoyudmzxf...,bafyreiedvs72vpdksf5... --perm readwrite

# ...or every space (dynamic, includes future spaces)
anytype auth apikey create mcp --scope jsonapi --all-spaces --perm read

# Change the grant later, in place: same key string, no client re-config;
# heart drops its session cache so the new grant applies immediately
anytype auth apikey grant <appHash> --spaces <id,id,...> --perm read
anytype auth apikey grant <appHash> --clear     # back to every-space access

# Scope and grant now visible in the table
anytype auth apikey list
```

Rules inherited from heart: grants exist only on `JsonAPI` keys (`Limited`
keys cannot carry one; `Full` is reserved for account-key sessions and cannot
be minted); `--spaces` and `--all-spaces` are mutually exclusive.

## Publishing

`.github/workflows/release-api-v2.yml` is a manual (workflow_dispatch) build:
static linux amd64 + arm64 musl binaries following the upstream alpine flow,
pushed as a multi-arch image to ghcr. The image tag is an input, defaulting
to `api-v2`:

```
ghcr.io/geeksquirrel/anytype-cli:api-v2
```

## Using it in any-sync-dockercompose

```yaml
# docker-compose.override.yml
services:
  anytype-cli:
    image: ghcr.io/geeksquirrel/anytype-cli:api-v2
  anytype-cli_bootstrap:
    image: ghcr.io/geeksquirrel/anytype-cli:api-v2
```

Then `docker compose pull anytype-cli anytype-cli_bootstrap && docker compose up -d anytype-cli`.

> The first ghcr package pushed with GITHUB_TOKEN is **private** by default.
> Switch it to Public under GitHub → Packages → anytype-cli → Package
> settings, or `docker login ghcr.io` on the deployment host before pulling.

## The legacy mcp-enhance line

Before v2 existed, this fork shipped the **mcp-enhance** patch set
(`patches/`, `release-mcp-enhance.yml`) on top of heart v0.50.20 to bring
discussion and rich-markdown capabilities to the v1 API. The v2 API supersedes
its main features natively (object `discussion` field, discussion cold-start,
chat message reading); the patch line still exists for v1-era deployments and
is retired the same way once upstream moves on.

## Retiring once upstream catches up

1. Watch upstream anytype-cli releases; when a release pins
   `anytype-heart >= v0.51.0`, its image serves v2 natively;
2. Delete `release-api-v2.yml` (and the legacy `release-mcp-enhance.yml` /
   `patches/`), restore `release.yml` to the upstream version, and point
   `docker-compose.override.yml` back at the official image.
