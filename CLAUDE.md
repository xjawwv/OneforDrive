# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RouteStorage is a cloud storage router that pools multiple Google Drive accounts into one virtual storage pool. Files are split into 256 MB chunks, each routed to whichever connected Drive account has the most free space. Chunk→Drive mapping is stored in MySQL. On download, chunks are fetched in index order and streamed as a merged file.

## Development Commands

```bash
# Start all services (MySQL, Redis, backend, frontend)
docker compose up

# Frontend dev (runs outside Docker)
cd frontend && npm run dev    # http://localhost:3000

# Backend runs on port 8081 (Docker) or 8080 (local)

# Database CLI (interactive menu)
pip install rich   # first time only
python db.py       # launches interactive TUI
python db.py status
```

## Critical Constraints

**Frontend:**
- Node 20 only (`node:20-alpine`). Node 22/24 break `npm install`.
- `npm install` requires `--legacy-peer-deps` (hardcoded in `.npmrc`). Never remove it.
- `lucide-vue-next@0.395.0` only — the package is deprecated but this version is confirmed working. Do not upgrade without verifying on npm registry first.
- **No `@apply` with CSS variable values** — Tailwind build fails silently. Write plain CSS for anything using `var(--color-*)`. Only `@apply` standard utilities like `flex`, `items-center`.
- **`focus:ring-[var(...)]/20` is illegal in `@apply`** — opacity modifier on arbitrary values crashes PostCSS. Use `box-shadow` in plain CSS instead.

**Backend:**
- Go 1.22 only. `go.sum` is generated at Docker build time via `go mod tidy` — never commit it.
- No `version:` field in `docker-compose.yml` (Compose v2 — causes warnings).

## Architecture

**Monorepo** — `frontend/` (Nuxt 3 SPA) + `backend/` (Go 1.22/Gin) + `docker-compose.yml` + `db.py` (Python CLI).

### Frontend (Nuxt 3 SPA — SSR disabled)

| Path | Purpose |
|---|---|
| `pages/` | Route pages — render only content via `<slot />`, layout handled by `layouts/default.vue` |
| `layouts/default.vue` | Owns AppSidebar + AppTopBar, provides `sidebarOpen` + topbar config via provide/inject |
| `components/AppSidebar.vue`, `AppTopBar.vue` | Layout components (owned by default.vue) |
| `composables/useApi.ts` | JWT-injecting `$fetch` wrapper — all API calls go through `apiFetch()` |
| `composables/usePermissions.ts` | RBAC state — `can(permKey)` checks permission list |
| `composables/useFeatureRoute.ts` | Feature flag toggling per route |
| `assets/css/tokens.css` | Design token definitions (light + dark themes via `[data-theme]`) |
| `assets/css/main.css` | Tailwind + global component classes (`.card`, `.btn-primary`, `.input-field`) |

**Topbar pattern:** Pages set topbar state via `useState('topbar', {...})`. Title/actions are injected via `provide/inject` (`topbar:title`, `topbar:actions` render functions).

**Pages without layout:** `login.vue`, `index.vue`, `shared/[token].vue` use `layout: false`.

### Backend (Go 1.22 / Gin)

```
cmd/main.go                     Entry point — MySQL/Redis init, route registration
internal/handler/               One file per domain: auth, account, file, file_upload,
                                file_download, share, storage, rbac, feature_route
internal/middleware/             JWT auth (extracts user_id) + RBAC permission checking
internal/model/                 Data models (User, DriveAccount, FileEntry, FileChunk, Role, Permission)
internal/repository/            MySQL queries + RBAC seed data + Redis caching
internal/service/google.go      Google Drive API (OAuth, token refresh, chunk routing)
pkg/redis/                      Redis client init
```

- Handlers receive DB via struct fields (`handler.FileHandler{DB: repository.DB}`).
- Auth middleware sets `user_id` on gin context. RBAC middleware checks permissions from Redis cache (5-minute TTL, key: `user_permissions:{user_id}`).
- DB schema auto-migrates on startup in `internal/repository/db.go`.

### Async Upload/Download Pattern

Uploads are async — backend responds immediately with file ID, then processes chunks in background goroutines. Frontend polls `/api/files/:id/progress`.

Downloads use session-based parallel processing — `POST /api/files/:id/download` starts an async session stored in an in-memory `sync.Map`. Frontend polls `/api/files/:id/download-progress`, then fetches the assembled file via `GET`.

## Design Conventions

- **Icons:** `lucide-vue-next` only — no emoji, no other icon libraries.
- **Colors:** always `var(--color-*)` tokens from `tokens.css` — never hardcode hex in Vue files.
- **Theme:** toggled via `data-theme` attribute, defined in `tokens.css` with dark mode overrides in `[data-theme="dark"]`.
- **Reused CSS classes:** `.card`, `.btn-primary`, `.input-field` defined in `main.css`.
- **Backend Go deps:** add to `go.mod` only — `go.sum` auto-generates on build.

## Google Drive Integration

- Google tokens refresh transparently via `internal/service/google.go` (checks `token_expiry`). Always use the provided token-refreshing client, not raw tokens.
- Orphan cleanup runs on account connect/sync — never call `CleanupOrphans` unless the user explicitly requests it.
- Shared links use 64-char hex tokens via `crypto/rand`. Public share routes (`/shared/:token`) do NOT require auth — never add auth middleware to share routes.

## MCP Servers

- **nuxt**: Use for any Nuxt API questions. `nuxt_list_documentation_pages` + `nuxt_get_documentation_page`.
- **context7**: Use for third-party library docs (Vue, Tailwind, Gin, etc.).
- **chrome-devtools**: Use for debugging, screenshots, Lighthouse audits, network analysis.
- **browsermcp**: Use when you need to interact with pages that require login/cookies.

## Git rules
- Claude may run `git status`, `git diff`, `git add`, `git commit`, and `git push`.
- Never run `git push --force`, `git reset --hard`, or `sudo`.
- Always show `git status` before commit.
- Use Conventional Commits.
- Never push secrets, `.env`, private keys, or credentials.
- Prefer pushing feature branches, not direct pushes to `main`.
