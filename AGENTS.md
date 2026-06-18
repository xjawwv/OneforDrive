# AGENTS.md

## Critical Gotchas (hard-won, not obvious from code)

1. **No `@apply` with CSS variable values** — Tailwind build fails silently.
   Write plain CSS for anything using `var(--color-*)`. Only `@apply` standard utilities like `flex`, `items-center`.

2. **`focus:ring-[var(...)]/20` is illegal in `@apply`** — opacity modifier `/20` on arbitrary values crashes PostCSS. Use `box-shadow` in plain CSS instead.

3. **Node version is `node:20-alpine` only** — `node:22` and `node:24` break `npm install` in this project.

4. **`npm install` requires `--legacy-peer-deps`** — hardcoded in `.npmrc`. Never remove it.

5. **No `go.sum` in repo** — generated at Docker build time via `go mod tidy`. Do not create or commit it.

6. **No `version:` field in `docker-compose.yml`** — obsolete in Compose v2, causes warnings.

7. **`lucide-vue-next@0.395.0` only** — package is deprecated but confirmed working. `@lucide/vue` version `0.477.0` does NOT exist on npm. Do not upgrade without verifying on registry first.

## Conventions (condensed)

- Icons: `lucide-vue-next` only — no emoji, no other icon libraries, ever
- Colors: always `var(--color-*)` tokens — never hardcode hex in Vue files
- Theme: toggled via `data-theme` attribute, defined in `tokens.css`
- Pages: each page owns its full layout (`definePageMeta({ layout: false })`)
- New Go deps: add to `go.mod` only — `go.sum` auto-generates on build

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

## Workflow for every task

1. Restate the task
2. **Check MCP servers** — which ones are relevant? Use them.
3. List files to change
4. Implement
5. **Verify with MCP** — screenshot the result, run a Lighthouse audit, check docs for correctness
6. Self-check: no emoji · no hardcoded colors · no `@apply` with CSS vars · lucide imports correct
7. Summarize changes
