# AGENTS.md

## Project Overview

**RouteStorage** is a cloud storage router. Files uploaded are split into 256 MB chunks, each routed to whichever connected Google Drive account has the most free space. Chunk-to-Drive mapping is stored in MySQL. On download, chunks are fetched in index order and streamed as a single merged file. This pools multiple Google Drive accounts into one virtual storage pool.

**Stack:** Nuxt 3 + Vue 3 + TypeScript + Tailwind CSS (frontend) · Go 1.22 + Gin (backend) · MySQL 8 + Redis 7 (data) · Docker Compose (infra) · Google Drive API v3 (storage)

## Folder Structure

```
backend/
  cmd/main.go                    Entry point — MySQL/Redis init, route registration
  internal/handler/              HTTP handlers (auth, account, file, file_upload, file_download, share, storage, rbac)
  internal/middleware/            JWT auth + RBAC permission checking
  internal/model/                Data models (User, DriveAccount, FileEntry, FileChunk, Role, Permission)
  internal/repository/           MySQL queries + RBAC seed data + Redis caching
  internal/service/google.go     Google Drive API integration (OAuth, token refresh, chunk routing)
  pkg/redis/                     Redis client init
frontend/
  pages/                         Each page owns its full layout (layout: false)
    index.vue                    Root redirect (auth check)
    login.vue                    Login/register
    explorer.vue                 Main file manager (~2350 lines)
    settings.vue                 Drive account management
    shared/[token].vue           Public share viewer
    admin/users.vue              User management
    admin/roles.vue              Role/permission management
  components/                    AppTopBar, AppSidebar
  composables/                   useApi (JWT fetch wrapper), usePermissions (RBAC state)
  assets/css/                    tokens.css (design tokens), main.css (Tailwind + globals)
db.py                            Python CLI for DB admin (uses docker compose exec)
```

## Critical Gotchas (hard-won, not obvious from code)

1. **No `@apply` with CSS variable values** — Tailwind build fails silently.
   Write plain CSS for anything using `var(--color-*)`. Only `@apply` standard utilities like `flex`, `items-center`.

2. **`focus:ring-[var(...)]/20` is illegal in `@apply`** — opacity modifier `/20` on arbitrary values crashes PostCSS. Use `box-shadow` in plain CSS instead.

3. **Node version is `node:20-alpine` only** — `node:22` and `node:24` break `npm install` in this project.

4. **`npm install` requires `--legacy-peer-deps`** — hardcoded in `.npmrc`. Never remove it.

5. **No `go.sum` in repo** — generated at Docker build time via `go mod tidy`. Do not create or commit it.

6. **No `version:` field in `docker-compose.yml`** — obsolete in Compose v2, causes warnings.

7. **`lucide-vue-next@0.395.0` only** — package is deprecated but confirmed working. `@lucide/vue` version `0.477.0` does NOT exist on npm. Do not upgrade without verifying on registry first.

8. **Uploads are async** — the backend responds immediately with the file ID after saving to a temp file, then processes chunks in background goroutines. The frontend polls `/api/files/:id/progress` for completion. Do not block on upload completion in the handler.

9. **Downloads use session-based parallel processing** — `POST /api/files/:id/download` starts an async session stored in an in-memory `sync.Map`. The frontend polls `/api/files/:id/download-progress`, then fetches the assembled file via `GET`. Do not use synchronous download for large files.

10. **RBAC permissions are cached in Redis for 5 minutes** — `internal/repository/rbac.go` caches permission lookups. If you modify roles/permissions, the cache won't reflect changes immediately. The cache key pattern is `user_permissions:{user_id}`.

11. **Orphan cleanup runs on account connect/sync** — `internal/service/google.go` deletes Drive files not tracked in the database. Never call `CleanupOrphans` unless the user explicitly requests it or during the OAuth callback flow.

12. **Google Drive tokens refresh transparently** — `internal/service/google.go` checks `token_expiry` and refreshes before API calls. If you add new Drive API calls, always use the provided token-refreshing client, not raw tokens.

13. **Every page owns its full layout** — all pages use `definePageMeta({ layout: false })` and import `AppSidebar` + `AppTopBar` directly. Do not use Nuxt's layout system.

14. **`explorer.vue` is the largest file (~2350 lines)** — contains the full file manager UI. When editing, be surgical. Prefer extracting logic into composables if adding significant new features.

15. **Shared links use 64-char hex tokens** — generated via `crypto/rand`. The public share routes (`/shared/:token`) do NOT require auth. Never add auth middleware to share routes.

## Conventions

### Frontend
- **Icons:** `lucide-vue-next` only — no emoji, no other icon libraries, ever
- **Colors:** always `var(--color-*)` tokens — never hardcode hex in Vue files
- **Theme:** toggled via `data-theme` attribute, defined in `tokens.css` with dark mode overrides in `[data-theme="dark"]`
- **Layouts:** each page owns its full layout (`definePageMeta({ layout: false })`)
- **Composables:** `useApi` for API calls (auto-injects JWT), `usePermissions` for RBAC checks
- **CSS classes:** `.card`, `.btn-primary`, `.input-field` etc. defined in `main.css`

### Backend
- **New Go deps:** add to `go.mod` only — `go.sum` auto-generates on build
- **Handlers:** one file per domain in `internal/handler/`
- **Auth middleware:** extracts JWT Bearer token, sets `user_id` on gin context
- **RBAC middleware:** checks `user_id` permissions from Redis cache
- **DB schema:** auto-migrated on startup in `internal/repository/db.go`

## API Reference

### Auth
```
POST /api/auth/register          Create account (auto-assigns "member" role)
POST /api/auth/login             Login (returns JWT)
```

### Drive Accounts
```
GET  /api/accounts               List connected accounts
GET  /api/accounts/connect       Get Google OAuth URL
GET  /api/accounts/oauth/callback OAuth redirect handler (public)
POST /api/accounts/:id/sync      Sync capacity + cleanup orphans
DELETE /api/accounts/:id         Remove account
```

### Files
```
GET    /api/files                List files (query: parent_id, account_id)
GET    /api/files/breadcrumb     Get folder breadcrumb path
POST   /api/files/upload         Upload file (async chunked processing)
POST   /api/files/folder         Create folder
GET    /api/files/:id/download   Start download session (async)
GET    /api/files/:id/download-progress  Poll download progress
DELETE /api/files/download-cancel        Cancel download session
GET    /api/files/:id/progress   Poll upload progress
GET    /api/files/:id/info       Get file metadata
DELETE /api/files/:id            Delete file (cleans up Drive chunks)
POST   /api/files/:id/share      Create share link (with expiry)
GET    /api/files/:id/shares     List share links
DELETE /api/files/:id/share/:linkId  Revoke share link
GET    /api/files/:id/thumbnail  Proxy thumbnail from Drive
```

### Storage
```
GET /api/storage/stats           Aggregate storage usage
```

### RBAC
```
GET    /api/rbac/me/permissions         Get current user's permissions
GET    /api/rbac/roles                  List all roles
POST   /api/rbac/roles                  Create role
GET    /api/rbac/permissions            List all permissions
GET    /api/rbac/roles/:id/permissions  Get role's permissions
PUT    /api/rbac/roles/:id/permissions  Update role's permissions
GET    /api/rbac/users/:id/roles        Get user's roles
POST   /api/rbac/users/:id/roles        Assign role to user
DELETE /api/rbac/users/:id/roles/:role_id  Remove role from user
```

### Public Share Routes (no auth)
```
GET /shared/:token                 Access shared file/folder
GET /shared/:token/download        Download shared file
GET /shared/:token/download-all    Download all files (ZIP)
GET /shared/:token/thumbnail       Thumbnail for shared file
```

## Database Schema

```sql
users           (id, email, password_hash, name, quota_limit, quota_used)
drive_accounts  (id, user_id, email, access_token, refresh_token, token_expiry,
                 capacity_total, capacity_used, route_storage_folder_id, is_active)
files           (id, user_id, name, mime_type, size_total, status, upload_progress,
                 parent_id, is_folder)
file_chunks     (id, file_id, chunk_index, chunk_size, drive_file_id, account_id, checksum)
shared_links    (id, file_id, token, expires_at, created_by)
roles           (id, name, description, is_system)
permissions     (id, key, description, category)
role_permissions (role_id, permission_id)
user_roles      (user_id, role_id)
```

**Seeded roles:** owner (all perms), admin (all perms), member (drive_accounts.manage_own, files.upload, files.delete_own)

## Mandatory MCP Usage

You MUST use MCP servers for the following tasks. Do not guess, do not rely on training data alone.

### Debugging & Browser Inspection
- **chrome-devtools**: Use for performance audits, Lighthouse, network analysis, console errors, screenshots, and running JS in-page. Always use this before concluding a bug is "hard to reproduce."
- **browsermcp**: Use when you need to interact with a page that requires existing login/cookies. Connect via the Chrome extension first.

### Documentation Lookup
- **nuxt**: Use `nuxt_list_documentation_pages` and `nuxt_get_documentation_page` for ANY Nuxt question. Do not guess Nuxt APIs — fetch the docs.
- **context7**: Use `context7_resolve-library-id` + `context7_query-docs` for ANY third-party library (Prisma, Tailwind, Vue, etc.). Training data may be stale — always verify current docs.

### Git & GitHub Operations
- **git**: Use for local git operations (status, diff, commit, branch, log). Do not shell out to `git` directly when MCP tools are available.
- **github**: Use for ALL GitHub API operations — PRs, issues, code search, Actions workflows, security alerts. Do not use `gh` CLI when the MCP server provides structured data.

### MCP Server Priority
1. For Nuxt questions → `nuxt` MCP first, then `context7`
2. For other libraries → `context7` first
3. For debugging UI bugs → `chrome-devtools` screenshot + console, or `browsermcp` for logged-in pages
4. For performance issues → `chrome-devtools` performance trace + Lighthouse audit
5. For git operations → `git` MCP
6. For GitHub operations → `github` MCP

### Never Skip MCP When:
- Verifying library API syntax (use context7 or nuxt)
- Debugging a reported UI bug (use chrome-devtools or browsermcp)
- Checking performance (use chrome-devtools)
- Reviewing PRs or issues (use github)
- Checking git status before committing (use git)

### Mandatory Debugging After Code Changes
**After writing ANY code or running `docker compose`**, you MUST verify the result in a real browser. No exceptions.

1. **After frontend code changes** → Open the page in `browsermcp` or `chrome-devtools`, take a screenshot, check for console errors. If the page has issues, fix them before finishing.
2. **After `docker compose up` / backend changes** → Open the frontend URL in `browsermcp` or `chrome-devtools`, navigate through the affected flows, check console for errors, verify the UI renders correctly.
3. **After UI/UX changes** → Take a screenshot via `chrome-devtools` or `browsermcp`, visually inspect the result, check responsive behavior if applicable.
4. **After API/route changes** → Use `browsermcp` or `chrome-devtools` to test the flow end-to-end in the browser, not just curl.

**This is non-negotiable.** "It should work" is not acceptable — you must SEE it work. If debugging tools are unavailable, tell the user before proceeding.

## Workflow for every task

1. Restate the task
2. **Check MCP servers** — which ones are relevant? Use them.
3. List files to change
4. Implement
5. **Verify with MCP** — screenshot the result, run a Lighthouse audit, check docs for correctness
6. Self-check: no emoji · no hardcoded colors · no `@apply` with CSS vars · lucide imports correct
7. **Git commit & push** — stage changed files, commit with a clear message, and push to remote. Every update must be committed. Never leave uncommitted changes after completing a task.
8. Summarize changes

### Git Commit Rules
- **Always commit after completing a task** — no exceptions, even for small changes.
- **Commit message format**: `<type>: <description>` (e.g., `feat: add share link expiry`, `fix: resolve upload progress polling`)
- **Types**: `feat`, `fix`, `refactor`, `docs`, `chore`, `style`, `test`
- **Push after commit** — do not leave local-only commits.
- **Never commit secrets, tokens, or API keys** — check `git diff` before committing.
- **Use `git` MCP** for all git operations (status, add, commit, push, diff, log).
