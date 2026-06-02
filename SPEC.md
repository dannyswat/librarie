# Librarie — MVP Stage 1 Specification

## Overview

Librarie is a school-facing learning management platform that consolidates educational materials, enables teachers to create rich multimedia content, supports AI-assisted authoring, and provides tools for building and delivering assessments.

---

## Deployment Model

- Stage 1 is a **single-school deployment** per instance.

---

## User Roles

| Role | Description |
|------|-------------|
| `admin` | Manages school settings, teachers, invitations, AI provider credentials, and teacher-subject assignments. Can do everything a teacher can. |
| `teacher` | Creates and manages content, question banks, and assessments **only** under assigned subjects. |
| `student` | Exists as an account role in Stage 1, but has no functional capabilities yet. |

---

## Authentication

### Login methods
- **Username + password** — bcrypt-hashed passwords stored in Postgres.
- **Passkey (WebAuthn / FIDO2)** — device-bound credential; requires registration of at least one authenticator per account.

### Rate limiting
- Apply limits on both username and IP using token-bucket strategy.
- Bucket capacity: 5 failed attempts.
- Refill rate: 1 token per minute.
- Duplicate password trials (same username + same password fingerprint within active window) do not consume additional tokens.
- Return `429 Too Many Requests` with a `Retry-After` header.
- Implement at the Echo middleware layer backed by Redis or a Postgres-backed limiter table.
- Passkey authentication failures are rate-limited using the same strategy, scoped per IP (no username/credential-id bucket for passkey flows to avoid credential enumeration).

### Session persistence
- On successful login, issue a signed, opaque session token stored as an `HttpOnly`, `Secure`, `SameSite=Strict` cookie.
- Sessions stored in a `sessions` table in Postgres (id, user_id, token_hash, created_at, expires_at, last_seen_at).
- Session expires after 30 days of inactivity (`last_seen_at` based sliding window).
- Logout invalidates the session row.

### Registration
- **Invitation flow**: Admin sends invitation links for teacher onboarding only in Stage 1. Recipient registers with a chosen password and/or passkey. Invitation links expire after **7 days**.
- **Defined admin bootstrap**: A set of usernames listed in server configuration (env var or config file) are auto-created as admins on first startup, or auto-elevated on first login, without requiring an invitation. This is for initial bootstrapping only.

---

## Content Structure

```
School
└── Subject (e.g. "Mathematics", "History")
  ├── Topics (many-to-many; shared topics allowed)
  └── Contents
    └── Pages (ordered list)
        └── Content Blocks (ordered list)
```

- A **Subject** belongs to a school and has a name, description, and optional cover image.
- A **Topic** can belong to multiple subjects (many-to-many).
- A **Content** belongs to exactly one subject and one or more topics within that subject context.
- A **Page** is an ordered unit within a content item and contains an ordered list of content blocks.
- Teachers can manage only content attached to subjects assigned to them by admin.

---

## Content Blocks

Each page contains an ordered list of **content blocks**. A block has:

| Field | Description |
|-------|-------------|
| `id` | UUID |
| `page_id` | Parent page |
| `type` | Enum (see below) |
| `position` | Integer ordering index |
| `data` | JSONB payload specific to each type |
| `created_at` / `updated_at` | Timestamps |

### Block types

| Type | Description | Key `data` fields |
|------|-------------|-------------------|
| `text` | Rich formatted text (prose) | `content` (HTML/markdown) |
| `article` | Long-form article with a title *(inferred addition; not explicitly in requirements)* | `title`, `content` |
| `speech` | Audio content and speech workflows | `source_type` (`upload`/`recorded`/`tts`/`stt`), `audio_url`, `transcript`, `voice_profile`, `conversation` (optional) |
| `flash_card` | Interactive front/back card pair (teacher + student capable) | `front`, `back`, `interaction_mode` |
| `image` | Image by upload, external URL, or AI generation | `source_type` (`upload`/`url`/`ai_generated`), `storage_key`, `external_url`, `alt_text`, `caption`, `generation_prompt` |
| `diagram` | Diagram block backed by Excalidraw | `diagram_tool` (`excalidraw`), `diagram_json`, `preview_image_key`, `caption` |
| `video` | Video by embed URL, upload, or AI generation | `source_type` (`embed_url`/`upload`/`ai_generated`), `embed_url`, `storage_key`, `caption`, `generation_prompt` |
| `translation` | Single source text with multiple translations | `source_language`, `source`, `translations[]` (language + text + audience tags) |

Flash card runtime is modeled for both teachers and students; in Stage 1, interaction is available in teacher flows, while student-side delivery remains disabled with other student functionality.

---

## AI Assistance

Teachers can invoke AI assistance at any point during content authoring:

- **Generate block content** — given a prompt and optional context (subject, topic), the AI returns a suggested content block payload.
- **Expand / rewrite** — given an existing block's content, request a rewrite, simplification, or expansion.
- **Suggest questions** — given page content, generate candidate assessment questions.
- **Multimodal generation** — support AI text explanation, audio generation, image generation, and video generation.
- **Speech experiences** — support TTS/STT flows and optional multi-voice AI conversation generation in `speech` blocks.

AI must be provider-pluggable:

- Define a provider interface in backend (`GenerateText`, `GenerateAudio`, `GenerateImage`, `GenerateVideo`, `TranscribeAudio`, `SynthesizeSpeech`).
- Implement adapter modules per provider (e.g. OpenAI, Anthropic + external media model providers).
- Store provider credentials encrypted at rest and managed by admin in settings UI.
- Select provider/model per capability with fallback strategy.
- No provider credentials are exposed to frontend.

---

## Question Bank

- Subject-wide question bank, with each question linked to one subject and one or more associated topics.
- A **question** has:
  - A **stem** — combinations such as text-only, image+text, audio+text, or video+text.
  - A **type** — one of the assessment block types below.
  - **Options / answer data** — type-specific payload stored as JSONB.
  - **Tags** — subject, topic, difficulty, custom labels.

### Assessment block (question) types

| Type | Description |
|------|-------------|
| `mc` | Multiple choice — one correct answer from 2–6 options |
| `true_false` | True / False statement |
| `writing` | Open-ended written response with optional rubric |
| `free_text` | Short free-text answer (single line) with optional model answer |
| `matching` | Match items in column A to items in column B |
| `image_mc` | Multiple choice where options are images |

Questions can be reused across multiple assessments.

---

## Assessments

- An **assessment** belongs to a subject/topic and contains an ordered list of questions drawn from the question bank.
- An assessment has: title, description, time limit (optional), passing score (optional), instructions.
- Teachers can add, remove, and reorder questions within an assessment.
- In Stage 1, assessments are for creation and management only; student participation/submission is out of scope.

---

## Presentation Mode

A read-only, distraction-free view of a page's content blocks, designed for classroom projection. Features:

- Full-screen toggle.
- Navigate between pages within a topic using keyboard arrows.
- Content blocks rendered without editing chrome.
- This mode is explicitly a clean display mode (not live collaboration).

---

## Technical Architecture

### Stack

| Layer | Technology |
|-------|------------|
| Backend API | Go, [Echo](https://echo.labstack.com/) |
| Database | PostgreSQL |
| ORM / queries | `sqlc` + raw SQL migrations |
| Frontend | React, TypeScript |
| Containerisation | Docker (backend + frontend only; Postgres runs externally) |
| File Storage | Local disk in Stage 1 via storage abstraction (S3-compatible extension point) |

### Backend structure (suggested)

```
/
├── cmd/
│   └── server/          # main entrypoint
├── internal/
│   ├── auth/            # session, passkey, password, rate limit
│   ├── content/         # subjects, topics, contents, pages, blocks
│   ├── assessment/      # questions, question bank, assessments
│   ├── ai/              # provider abstraction + adapters
│   ├── storage/         # local disk driver + S3-compatible interface
│   └── user/            # user management, invitations
├── db/
│   ├── migrations/      # SQL migration files
│   └── queries/         # sqlc query files
├── Dockerfile
└── docker-compose.yml   # backend + frontend services only
```

### Frontend structure (suggested)

```
src/
├── features/
│   ├── auth/
│   ├── admin/           # teachers, subject assignment, AI provider config
│   ├── content/
│   ├── assessment/
│   └── ai/
├── components/          # shared UI components
├── hooks/
├── api/                 # typed fetch wrappers
└── main.tsx
```

### API design principles

- RESTful JSON API under `/api/v1/`.
- Authentication via session cookie on all protected routes.
- Input validation at Echo middleware layer; structured error responses `{ "error": "...", "field_errors": {...} }`.
- File uploads via `multipart/form-data`; files stored on local disk in Stage 1 through a pluggable storage interface.

### Key API surface (indicative, not exhaustive)

```
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
POST   /api/v1/auth/passkey/register/begin
POST   /api/v1/auth/passkey/register/complete
POST   /api/v1/auth/passkey/authenticate/begin
POST   /api/v1/auth/passkey/authenticate/complete
POST   /api/v1/invitations
POST   /api/v1/invitations/:token/accept

GET    /api/v1/admin/teachers
POST   /api/v1/admin/teachers/invite
PUT    /api/v1/admin/teachers/:id/subjects
GET    /api/v1/admin/ai/providers/:provider_key/capabilities
PUT    /api/v1/admin/ai/providers/:provider_key/capabilities/:capability

GET    /api/v1/subjects
POST   /api/v1/subjects
GET    /api/v1/subjects/:id
PUT    /api/v1/subjects/:id
DELETE /api/v1/subjects/:id
GET    /api/v1/subjects/:id/topics
POST   /api/v1/subjects/:id/topics
POST   /api/v1/contents
GET    /api/v1/contents/:id
PUT    /api/v1/contents/:id
DELETE /api/v1/contents/:id
PUT    /api/v1/contents/:id/topics
GET    /api/v1/contents/:id/pages
POST   /api/v1/contents/:id/pages
GET    /api/v1/pages/:id
PATCH  /api/v1/pages/:id
DELETE /api/v1/pages/:id
PUT    /api/v1/pages/:id/blocks    # replace ordered block list

GET    /api/v1/questions
POST   /api/v1/questions
PUT    /api/v1/questions/:id
DELETE /api/v1/questions/:id
GET    /api/v1/assessments
POST   /api/v1/assessments
GET    /api/v1/assessments/:id
PUT    /api/v1/assessments/:id
DELETE /api/v1/assessments/:id

POST   /api/v1/ai/generate-block
POST   /api/v1/ai/suggest-questions
POST   /api/v1/ai/generate-audio
POST   /api/v1/ai/generate-image
POST   /api/v1/ai/generate-video
```

---

## Database schema (high-level)

```sql
users (id, username, email, password_hash, role, created_at)
passkeys (id, user_id, credential_id, public_key, sign_count, created_at)
sessions (id, user_id, token_hash, expires_at, last_seen_at, created_at)
login_attempts (id, username, ip, password_fingerprint, attempted_at, success)
login_rate_limits (scope_type, scope_key, tokens, updated_at)

invitations (id, email, role, token_hash, invited_by, expires_at, accepted_at)

subjects (id, name, description, cover_image_key, position, created_by, created_at)
teachers_subjects (teacher_id, subject_id, assigned_by, assigned_at)
topics (id, name, description, created_by, created_at)
subject_topics (subject_id, topic_id, position)
contents (id, subject_id, title, description, position, created_by, created_at, updated_at)
content_topics (content_id, topic_id)
pages (id, content_id, name, position, created_by, created_at, updated_at)
blocks (id, page_id, type, position, data jsonb, created_at, updated_at)

questions (id, subject_id, stem_type, stem_data jsonb, type, answer_data jsonb, tags text[], created_by, created_at)
question_topics (question_id, topic_id)
assessments (id, title, description, subject_id, time_limit_seconds, passing_score, created_by, created_at)
assessment_topics (assessment_id, topic_id)
assessment_questions (assessment_id, question_id, position)

ai_provider_configs (id, provider_key, capability, encrypted_credentials, model, is_enabled, updated_at)
-- storage driver is configured via environment variable / config file; no runtime DB table needed
```

---

## Out of scope for Stage 1

- Student functionality (student role exists but no learner features yet).
- Assessment taking/submission by students.
- Grading and result analytics.
- Real-time collaboration on content.
- Notifications / email sending (beyond invitation emails).
- Password reset / forgot-password flow.
- Mobile applications.
- LMS integrations (Google Classroom, Canvas, etc.).
- S3-compatible remote storage (storage abstraction is implemented; remote driver is wired in a later stage).
