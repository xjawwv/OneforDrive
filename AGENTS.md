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

## Workflow for every task

1. Restate the task
2. List files to change
3. Implement
4. Self-check: no emoji · no hardcoded colors · no `@apply` with CSS vars · lucide imports correct
5. Summarize changes
