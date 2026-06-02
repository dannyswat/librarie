-- ───────────────────────── Subjects ─────────────────────────

-- name: CreateSubject :one
INSERT INTO subjects (name, description, cover_image_key, position, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSubjectByID :one
SELECT * FROM subjects WHERE id = $1;

-- name: ListSubjects :many
SELECT * FROM subjects ORDER BY position ASC, created_at ASC;

-- name: UpdateSubject :one
UPDATE subjects
SET name = $2, description = $3, cover_image_key = $4, position = $5
WHERE id = $1
RETURNING *;

-- name: DeleteSubject :exec
DELETE FROM subjects WHERE id = $1;

-- ───────────────────────── Teachers ↔ Subjects ─────────────────────────

-- name: AssignTeacherToSubject :exec
INSERT INTO teachers_subjects (teacher_id, subject_id, assigned_by)
VALUES ($1, $2, $3)
ON CONFLICT (teacher_id, subject_id) DO NOTHING;

-- name: UnassignTeacherFromSubject :exec
DELETE FROM teachers_subjects WHERE teacher_id = $1 AND subject_id = $2;

-- name: ListSubjectsByTeacher :many
SELECT s.* FROM subjects s
JOIN teachers_subjects ts ON ts.subject_id = s.id
WHERE ts.teacher_id = $1
ORDER BY s.position ASC;

-- name: ListTeachersBySubject :many
SELECT u.* FROM users u
JOIN teachers_subjects ts ON ts.teacher_id = u.id
WHERE ts.subject_id = $1
ORDER BY ts.assigned_at ASC;

-- ───────────────────────── Topics ─────────────────────────

-- name: CreateTopic :one
INSERT INTO topics (name, description, created_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTopicByID :one
SELECT * FROM topics WHERE id = $1;

-- name: ListTopics :many
SELECT * FROM topics ORDER BY name ASC;

-- name: UpdateTopic :one
UPDATE topics SET name = $2, description = $3 WHERE id = $1 RETURNING *;

-- name: DeleteTopic :exec
DELETE FROM topics WHERE id = $1;

-- ───────────────────────── Subjects ↔ Topics ─────────────────────────

-- name: AddTopicToSubject :exec
INSERT INTO subject_topics (subject_id, topic_id, position)
VALUES ($1, $2, $3)
ON CONFLICT (subject_id, topic_id) DO NOTHING;

-- name: RemoveTopicFromSubject :exec
DELETE FROM subject_topics WHERE subject_id = $1 AND topic_id = $2;

-- name: ListTopicsBySubject :many
SELECT t.* FROM topics t
JOIN subject_topics st ON st.topic_id = t.id
WHERE st.subject_id = $1
ORDER BY st.position ASC;

-- ───────────────────────── Contents ─────────────────────────

-- name: CreateContent :one
INSERT INTO contents (subject_id, title, description, position, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetContentByID :one
SELECT * FROM contents WHERE id = $1;

-- name: ListContentsBySubject :many
SELECT * FROM contents WHERE subject_id = $1 ORDER BY position ASC;

-- name: UpdateContent :one
UPDATE contents
SET title = $2, description = $3, position = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteContent :exec
DELETE FROM contents WHERE id = $1;

-- ───────────────────────── Contents ↔ Topics ─────────────────────────

-- name: AddTopicToContent :exec
INSERT INTO content_topics (content_id, topic_id)
VALUES ($1, $2)
ON CONFLICT (content_id, topic_id) DO NOTHING;

-- name: RemoveTopicFromContent :exec
DELETE FROM content_topics WHERE content_id = $1 AND topic_id = $2;

-- name: ReplaceContentTopics :exec
DELETE FROM content_topics WHERE content_id = $1;

-- name: ListTopicsByContent :many
SELECT t.* FROM topics t
JOIN content_topics ct ON ct.topic_id = t.id
WHERE ct.content_id = $1
ORDER BY t.name ASC;

-- name: ListContentsByTopic :many
SELECT c.* FROM contents c
JOIN content_topics ct ON ct.content_id = c.id
WHERE ct.topic_id = $1
ORDER BY c.position ASC;

-- ───────────────────────── Pages ─────────────────────────

-- name: CreatePage :one
INSERT INTO pages (content_id, name, position, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPageByID :one
SELECT * FROM pages WHERE id = $1;

-- name: ListPagesByContent :many
SELECT * FROM pages WHERE content_id = $1 ORDER BY position ASC;

-- name: UpdatePage :one
UPDATE pages
SET name = $2, position = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePage :exec
DELETE FROM pages WHERE id = $1;

-- ───────────────────────── Blocks ─────────────────────────

-- name: CreateBlock :one
INSERT INTO blocks (page_id, type, position, data)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetBlockByID :one
SELECT * FROM blocks WHERE id = $1;

-- name: ListBlocksByPage :many
SELECT * FROM blocks WHERE page_id = $1 ORDER BY position ASC;

-- name: UpdateBlock :one
UPDATE blocks
SET type = $2, position = $3, data = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteBlock :exec
DELETE FROM blocks WHERE id = $1;

-- name: DeleteBlocksByPage :exec
DELETE FROM blocks WHERE page_id = $1;
