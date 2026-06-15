# RouteStorage — Specification

## Stack
| Layer | Tech |
|---|---|
| Frontend | Nuxt 3.13.2, Vue 3.5.13, TypeScript, Tailwind 3.4.14 |
| Icons | lucide-vue-next 0.395.0 |
| Backend | Golang 1.22, Gin |
| DB | MySQL 8, Redis 7 |
| Infra | Docker Compose, node:20-alpine |

## How it works
Files uploaded to RouteStorage are split into 256MB chunks. Each chunk is routed to whichever connected Google Drive account has enough free space. Chunk→Drive mapping is stored in MySQL. On download, chunks are fetched in index order and streamed as a merged file.

## Database schema
```sql
users           (id, email, password_hash, name, quota_limit, quota_used)
drive_accounts  (id, user_id, email, access_token, refresh_token, token_expiry, capacity_total, capacity_used, is_active)
files           (id, user_id, name, mime_type, size_total, status, parent_id, is_folder)
file_chunks     (id, file_id, chunk_index, chunk_size, drive_file_id, account_id, checksum)
```

## API endpoints
```
POST   /api/auth/register
POST   /api/auth/login
GET    /api/accounts
GET    /api/accounts/connect
GET    /api/accounts/oauth/callback
DELETE /api/accounts/:id
GET    /api/files?parent_id=&account_id=
POST   /api/files/upload
GET    /api/files/:id/download
DELETE /api/files/:id
GET    /api/files/:id/info
POST   /api/files/folder
GET    /api/storage/stats
```

## Environment variables
```env
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
REDIS_HOST, REDIS_PORT
JWT_SECRET
GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET
GOOGLE_REDIRECT_URL=http://localhost:8080/api/accounts/oauth/callback
CHUNK_SIZE_MB=256
NUXT_PUBLIC_API_BASE=http://localhost:8080
```