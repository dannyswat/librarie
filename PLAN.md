# Librarie — MVP Stage 1 Implementation Plan

## Guiding Principles

- Build vertically (thin end-to-end slices) within each phase so the system is runnable after every phase.
- Backend and frontend tasks within a phase can proceed in parallel once the API contract is agreed.
- Each phase ends with a working, deployable build.

---

## Phase 0 — Project Scaffolding

Goal: empty-but-runnable backend and frontend, wired together in Docker Compose.

### 0.1 Repository structure
- Initialise Go module (`go mod init`).
- Create directory skeleton per SPEC (`cmd/server`, `internal/…`, `db/migrations`, `db/queries`).
- Initialise React/TypeScript project with Vite under `web/`.
- Add `.gitignore`, `README.md`, `Makefile` with common targets (`dev`, `build`, `migrate`, `sqlc-gen`, `test`).

### 0.2 Backend bootstrap
- Add Echo server entry point in `cmd/server/main.go` with configurable port.
- Add structured logger middleware (zerolog or slog).
- Add health-check endpoint `GET /health`.
- Add environment/config loading (`LIBRARIE_*` env vars via a config struct).
- Add `admin_bootstrap_users` config key (comma-separated usernames).

### 0.3 Database tooling
- Choose and wire migration tool (golang-migrate or goose).
- Add `db/migrations/000_init.sql` placeholder and verify it runs.
- Configure sqlc (`sqlc.yaml`) pointing at `db/queries/` and `db/migrations/`.

### 0.4 Frontend bootstrap
- Vite + React + TypeScript with path aliases (`@/`).
- Add React Router for SPA routing.
- Add Axios (or `fetch` wrapper) base client with cookie credentials.
- Proxy `/api` to backend in Vite dev config.
- Add placeholder `App.tsx` rendering a "Librarie" heading.

### 0.5 Docker Compose
- `Dockerfile` for backend (multi-stage: build → minimal runtime image).
- `Dockerfile` for frontend (build → nginx static serve).
- `docker-compose.yml`: backend + frontend services; Postgres runs externally (env var `DATABASE_URL`).
- Document local dev flow in README.

**Deliverable:** `docker compose up` serves the health endpoint and the React placeholder page.

---

## Phase 1 — Full Database Schema

Goal: all tables created via versioned migrations; sqlc queries generated for every table.

### 1.1 Migrations (one file per logical group)
- `001_users_auth.sql` — `users`, `passkeys`, `sessions`, `login_attempts`, `login_rate_limits`
- `002_invitations.sql` — `invitations`
- `003_content.sql` — `subjects`, `teachers_subjects`, `topics`, `subject_topics`, `contents`, `content_topics`, `pages`, `blocks`
- `004_questions_assessments.sql` — `questions`, `question_topics`, `assessments`, `assessment_topics`, `assessment_questions`
- `005_ai_config.sql` — `ai_provider_configs`

Apply column constraints:
- All `id` columns: `uuid DEFAULT gen_random_uuid() PRIMARY KEY`
- All FK columns: explicit `REFERENCES … ON DELETE CASCADE` or `RESTRICT` as appropriate.
- `blocks.data`, `questions.stem_data`, `questions.answer_data`: `jsonb NOT NULL DEFAULT '{}'`
- `users.role`: `text NOT NULL CHECK (role IN ('admin','teacher','student'))`
- `login_rate_limits`: composite primary key `(scope_type, scope_key)`

### 1.2 sqlc query stubs
- Write named queries (SELECT, INSERT, UPDATE, DELETE) for every table.
- Run `sqlc generate` and verify compilation.
- No business logic yet — queries only.

**Deliverable:** `make migrate` and `make sqlc-gen` both succeed cleanly.

---

## Phase 2 — Authentication

Dependencies: Phase 0, Phase 1.

### 2.1 Internal packages
- `internal/auth/password.go` — bcrypt hash + verify helpers.
- `internal/auth/session.go` — generate opaque token, hash for storage, set/clear cookie helpers.
- `internal/auth/middleware.go` — Echo middleware that reads session cookie, loads user from DB, attaches to context; returns `401` if missing/expired; updates `last_seen_at`.

### 2.2 Admin bootstrap
- On server startup, read `LIBRARIE_ADMIN_USERS` config.
- For each username: upsert user with `role = 'admin'`; set a random unusable password hash if the row is new.

### 2.3 Password login
- `POST /api/v1/auth/login` — validate body, lookup user, bcrypt compare, create session row, set cookie.
- Apply rate-limit middleware (see 2.5) before the handler.

### 2.4 Logout
- `POST /api/v1/auth/logout` — delete session row, clear cookie.

### 2.5 Rate limiting
- `internal/auth/ratelimit.go` — token-bucket logic backed by `login_rate_limits` table.
  - Scopes: `username:<value>` and `ip:<value>`.
  - Capacity: 5, refill: 1/min.
  - Dedup: compute `sha256(username + ":" + password)` as fingerprint; skip bucket decrement if same fingerprint already recorded for that username in the current window.
- Echo middleware wrapping login and passkey authenticate endpoints.
- Return `429` with `Retry-After` header on exhaustion.

### 2.6 Passkey / WebAuthn
- Add `go-webauthn/webauthn` dependency.
- `internal/auth/passkey.go` — configure relying party, store/load challenge in session (or a short-lived DB row), credential helpers.
- `POST /api/v1/auth/passkey/register/begin` — generate registration options, return to client.
- `POST /api/v1/auth/passkey/register/complete` — verify attestation, store `passkeys` row.
- `POST /api/v1/auth/passkey/authenticate/begin` — generate assertion options.
- `POST /api/v1/auth/passkey/authenticate/complete` — verify assertion, update `sign_count`, create session.
- IP-scoped rate limiting on authenticate endpoints (reuse 2.5 middleware).

### 2.7 Invitation flow
- `POST /api/v1/invitations` (admin only) — generate `crypto/rand` token, store hash + 7-day expiry, send email (stub: log to stdout in Stage 1).
- `POST /api/v1/invitations/:token/accept` — verify token hash + expiry, create `users` row with provided password/passkey, mark invitation accepted.

### 2.8 Frontend — Auth screens
- `/login` page: username + password form; WebAuthn "sign in with passkey" button.
- `/invite/:token` page: registration form (username, password, optional passkey enrol).
- Auth context / hook (`useAuth`) storing current user; redirect logic on 401.
- Protected route wrapper.

**Deliverable:** Login, logout, and passkey flows work end-to-end. Admin bootstrap creates the first admin user.

---

## Phase 3 — Admin & User Management

Dependencies: Phase 2.

### 3.1 Role authorisation middleware
- `internal/auth/require_role.go` — Echo middleware factory `RequireRole("admin")`, `RequireRole("teacher")`.
- Apply to all subsequent route groups.

### 3.2 Teacher management APIs
- `GET  /api/v1/admin/teachers` — list users with `role = 'teacher'`.
- `POST /api/v1/admin/teachers/invite` — alias for invitation creation with `role = 'teacher'`.
- `PUT  /api/v1/admin/teachers/:id/subjects` — replace the set of subject assignments in `teachers_subjects`.

### 3.3 AI provider config APIs
- `GET /api/v1/admin/ai/providers/:provider_key/capabilities` — list capabilities for a provider key.
- `PUT /api/v1/admin/ai/providers/:provider_key/capabilities/:capability` — upsert `ai_provider_configs` row; encrypt credentials before storage (AES-GCM with key from env `LIBRARIE_ENCRYPTION_KEY`).
- Credentials never returned to frontend; response includes only `provider_key`, `capability`, `model`, `is_enabled`.

### 3.4 Frontend — Admin screens
- `/admin/teachers` — teacher list, invite button (opens modal), per-teacher subject assignment UI.
- `/admin/ai-providers` — per-capability provider/model/credential form.

**Deliverable:** Admin can invite teachers, assign subjects, and configure AI providers.

---

## Phase 4 — Content Management

Dependencies: Phase 3.

### 4.1 Storage abstraction
- Define `internal/storage/storage.go` interface: `Put(ctx, key, reader, size) error`, `Get(ctx, key) (io.ReadCloser, error)`, `Delete(ctx, key) error`, `URL(key) string`.
- Implement `internal/storage/local.go` — stores files under a configured base directory; `URL` returns a signed path via a backend-served route.
- Wire `GET /uploads/:key` in Echo (auth-gated) to stream files from local storage.

### 4.2 Subjects
- `GET    /api/v1/subjects` — list subjects (teachers see only assigned; admin sees all).
- `POST   /api/v1/subjects` — create (admin only); set `position` = max+1.
- `GET    /api/v1/subjects/:id` — single subject.
- `PUT    /api/v1/subjects/:id` — update name/description/cover/position (admin only).
- `DELETE /api/v1/subjects/:id` — soft or hard delete (admin only); cascade via FK.
- File upload for `cover_image_key` via multipart — delegate to storage abstraction.

### 4.3 Topics
- `GET  /api/v1/subjects/:id/topics` — list topics for a subject (via `subject_topics`).
- `POST /api/v1/subjects/:id/topics` — create topic and link to subject; or link existing topic by id.

### 4.4 Contents
- `POST   /api/v1/contents` — create content under a subject (teacher must own subject).
- `GET    /api/v1/contents/:id` — fetch with topics.
- `PUT    /api/v1/contents/:id` — update title/description/position.
- `DELETE /api/v1/contents/:id` — delete (cascades pages + blocks).
- `PUT    /api/v1/contents/:id/topics` — replace topic associations.

### 4.5 Pages
- `GET    /api/v1/contents/:id/pages` — ordered list of pages.
- `POST   /api/v1/contents/:id/pages` — create page; position = max+1.
- `GET    /api/v1/pages/:id` — page with blocks.
- `PATCH  /api/v1/pages/:id` — rename or reorder (update `position`).
- `DELETE /api/v1/pages/:id` — delete page and its blocks.

### 4.6 Blocks
- `PUT /api/v1/pages/:id/blocks` — replace the full ordered block list (array of `{type, position, data}`); diff-and-upsert or delete-then-insert pattern.
- Validate `data` schema per `type` at the handler layer (use a type-switch with per-type struct validation).
- Handle file uploads embedded in block data (image `upload`, speech `upload`/`recorded`): receive file as multipart part, store via storage abstraction, replace payload field with `storage_key`.

Block type validation structs to implement:
- `text`, `article` — require non-empty `content`
- `speech` — validate `source_type` enum; require `audio_url` or `storage_key` for upload/recorded types
- `flash_card` — require `front`, `back`, `interaction_mode`
- `image` — validate `source_type`; require `storage_key` or `external_url` per type
- `diagram` — require `diagram_json`
- `video` — validate `source_type`; require `embed_url` or `storage_key`
- `translation` — require `source_language`, `source`, non-empty `translations[]`

### 4.7 Frontend — Content authoring
- Sidebar: subject list → content list → page list.
- Page editor: drag-and-drop ordered block list.
- Block palette: add block by type.
- Per-type block editors:
  - `text` / `article` — rich text editor (e.g. Tiptap).
  - `speech` — upload / record / TTS / STT tabs; conversation builder UI.
  - `flash_card` — front/back editor; flip preview.
  - `image` — upload / URL / AI generate tabs.
  - `diagram` — embedded Excalidraw iframe/component.
  - `video` — embed URL / upload / AI generate tabs.
  - `translation` — source text + add-language panel.
- Auto-save or explicit save that calls `PUT /api/v1/pages/:id/blocks`.

**Deliverable:** Teachers can create subjects, topics, contents, pages, and author all block types.

---

## Phase 5 — Question Bank & Assessments

Dependencies: Phase 4 (subject context, storage for media stems).

### 5.1 Questions
- `GET    /api/v1/questions` — list with filters: `subject_id`, `topic_id`, `type`, `tags`.
- `POST   /api/v1/questions` — create; validate `stem_type` (text / image+text / audio+text / video+text) and `type` enum; accept media uploads for stems.
- `PUT    /api/v1/questions/:id` — update.
- `DELETE /api/v1/questions/:id` — delete (unlinks from assessments via cascade or error if referenced).

Implement per-type answer data validation:
- `mc`, `image_mc` — require `options[]` (2–6), `correct_index`
- `true_false` — require `correct: bool`
- `writing` — optional `rubric`
- `free_text` — optional `model_answer`
- `matching` — require `column_a[]`, `column_b[]`, `pairs[]`

### 5.2 Assessments
- `GET    /api/v1/assessments` — list by subject.
- `POST   /api/v1/assessments` — create with title, description, optional time limit and passing score.
- `GET    /api/v1/assessments/:id` — fetch with ordered questions.
- `PUT    /api/v1/assessments/:id` — update metadata or reorder/add/remove questions (replace `assessment_questions` array).
- `DELETE /api/v1/assessments/:id` — delete.

### 5.3 Frontend — Question bank & assessment screens
- `/subjects/:id/questions` — filterable question bank list; create/edit question modal with per-type form.
- `/subjects/:id/assessments` — assessment list; assessment builder: search/add questions from bank, drag to reorder.

**Deliverable:** Teachers can build a question bank and compose assessments.

---

## Phase 6 — AI Integration

Dependencies: Phase 3 (provider config), Phase 4 (blocks), Phase 5 (questions).

### 6.1 Provider abstraction
- Define `internal/ai/provider.go` interface:
  ```go
  type Provider interface {
      GenerateText(ctx, prompt, opts) (string, error)
      GenerateAudio(ctx, text, voice, opts) (io.Reader, error)
      GenerateImage(ctx, prompt, opts) (io.Reader, error)
      GenerateVideo(ctx, prompt, opts) (io.Reader, error)
      TranscribeAudio(ctx, audio io.Reader, opts) (string, error)
      SynthesizeSpeech(ctx, text, voice, opts) (io.Reader, error)
  }
  ```
- `internal/ai/registry.go` — map `provider_key → Provider` loaded from `ai_provider_configs`; lazy-init per capability; decrypt credentials at load time.
- Fallback: if primary provider returns error, try next enabled provider for the same capability.

### 6.2 Provider adapters
- `internal/ai/openai/` — implement all six interface methods against OpenAI APIs (gpt-4o, tts-1, dall-e-3, whisper-1).
- `internal/ai/anthropic/` — implement `GenerateText` against Claude API.
- Stub remaining methods on Anthropic adapter to return `ErrNotSupported`.
- Design adapter registration so new providers can be added without touching core routing.

### 6.3 AI API endpoints
- All endpoints require authentication; teacher or admin role.
- Credentials are resolved server-side from `ai_provider_configs`; never passed from client.

| Endpoint | Input | Action |
|---|---|---|
| `POST /api/v1/ai/generate-block` | `{type, prompt, subject_id, topic_id}` | Return a suggested block `data` payload |
| `POST /api/v1/ai/suggest-questions` | `{page_id, count, types[]}` | Return array of candidate question payloads |
| `POST /api/v1/ai/generate-audio` | `{text, voice_profile}` | Generate audio; store via storage; return `storage_key` |
| `POST /api/v1/ai/generate-image` | `{prompt, style}` | Generate image; store via storage; return `storage_key` |
| `POST /api/v1/ai/generate-video` | `{prompt}` | Generate video; store via storage; return `storage_key` |

### 6.4 Frontend — AI authoring integration
- "Generate with AI" button in block palette and per-block toolbar.
- Prompt modal: enter prompt → call `generate-block` → preview result → insert or discard.
- "Suggest questions" button on page view → opens question preview panel → accept individual suggestions into the question bank.
- TTS / STT wired into `speech` block editor.
- Image / video generate tabs in `image` and `video` block editors.

**Deliverable:** Teachers can invoke AI assistance for all content and question types.

---

## Phase 7 — Presentation Mode

Dependencies: Phase 4.

### 7.1 Backend
- No new API endpoints needed; reuses `GET /api/v1/pages/:id`.
- Verify block read queries return all data needed for rendering.

### 7.2 Frontend
- Route `/present/:content_id/:page_id` (or `/present/:content_id` defaulting to page 1).
- Full-screen toggle (`document.documentElement.requestFullscreen()`).
- Keyboard navigation: `←`/`→` to move between pages within the content.
- Render each block in read-only mode (no edit chrome, no toolbars).
- Distraction-free CSS layout.

**Deliverable:** Teachers can present any page in full-screen classroom mode.

---

## Phase 8 — Hardening & Production Readiness

### 8.1 Security
- Audit all endpoints: confirm `RequireRole` middleware applied everywhere.
- Verify no credentials or internal errors leak in API responses (structured error filter middleware).
- Confirm `HttpOnly`/`Secure`/`SameSite=Strict` on session cookies in production config.
- Add `Content-Security-Policy`, `X-Frame-Options`, `X-Content-Type-Options` headers via Echo middleware.
- Ensure `LIBRARIE_ENCRYPTION_KEY` is required and non-empty at startup.
- Input size limits on file uploads (configurable max, default 50 MB).

### 8.2 Testing
- Unit tests for: password hashing, token-bucket rate limiter, invitation token generation/expiry, block data validators, AI provider registry fallback.
- Integration tests (using `testcontainers-go` or a test Postgres): auth flows, CRUD round-trips for content/questions/assessments.
- Frontend: component tests for block editors (Vitest + Testing Library); E2E smoke test for login → create content → present (Playwright).

### 8.3 Observability
- Structured JSON logs on all requests (method, path, status, latency, user_id).
- Expose `GET /metrics` (Prometheus counters for request count, error rate, AI provider calls).
- Graceful shutdown (drain in-flight requests on SIGTERM).

### 8.4 Developer experience
- `make dev` starts backend with hot-reload (air) and frontend with `vite dev`.
- `make test` runs all backend + frontend tests.
- `make migrate-create NAME=<name>` scaffolds a new migration file.
- Document all `LIBRARIE_*` environment variables in README.

---

## Dependency Graph

```
Phase 0 (Scaffolding)
    └── Phase 1 (Schema)
            └── Phase 2 (Auth)
                    └── Phase 3 (Admin)
                            ├── Phase 4 (Content)  ──────────────┐
                            │       └── Phase 5 (Questions)       │
                            │               └── Phase 6 (AI) ─────┤
                            │                                      │
                            └─────────────────────── Phase 7 (Presentation)
                                                            │
                                                    Phase 8 (Hardening)
```

---

## Key Technical Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Migration tool | `golang-migrate` | CLI + Go library; works well with raw SQL |
| WebAuthn library | `go-webauthn/webauthn` | Maintained, FIDO2-compliant |
| Rate-limit store | Postgres `login_rate_limits` table | No Redis dependency in Stage 1 |
| Credential encryption | AES-256-GCM, key from env | Standard, auditable; no external KMS needed in Stage 1 |
| Rich text | Tiptap (ProseMirror-based) | Extensible; works with React |
| Diagram | Excalidraw (embedded) | As specified |
| File storage | Local disk with S3-compatible interface | Extensible per spec |
| Frontend state | React Query + React Context | Minimal boilerplate; no Redux overhead for Stage 1 |
