-- ───────────────────────── Questions ─────────────────────────

-- name: CreateQuestion :one
INSERT INTO questions (subject_id, stem_type, stem_data, type, answer_data, tags, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: ListQuestionsBySubject :many
SELECT * FROM questions WHERE subject_id = $1 ORDER BY created_at DESC;

-- name: UpdateQuestion :one
UPDATE questions
SET stem_type = $2, stem_data = $3, type = $4, answer_data = $5, tags = $6
WHERE id = $1
RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;

-- ───────────────────────── Questions ↔ Topics ─────────────────────────

-- name: AddTopicToQuestion :exec
INSERT INTO question_topics (question_id, topic_id)
VALUES ($1, $2)
ON CONFLICT (question_id, topic_id) DO NOTHING;

-- name: RemoveTopicFromQuestion :exec
DELETE FROM question_topics WHERE question_id = $1 AND topic_id = $2;

-- name: ListTopicsByQuestion :many
SELECT t.* FROM topics t
JOIN question_topics qt ON qt.topic_id = t.id
WHERE qt.question_id = $1
ORDER BY t.name ASC;

-- name: ListQuestionsByTopic :many
SELECT q.* FROM questions q
JOIN question_topics qt ON qt.question_id = q.id
WHERE qt.topic_id = $1
ORDER BY q.created_at DESC;

-- ───────────────────────── Assessments ─────────────────────────

-- name: CreateAssessment :one
INSERT INTO assessments (title, description, subject_id, time_limit_seconds, passing_score, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAssessmentByID :one
SELECT * FROM assessments WHERE id = $1;

-- name: ListAssessmentsBySubject :many
SELECT * FROM assessments WHERE subject_id = $1 ORDER BY created_at DESC;

-- name: UpdateAssessment :one
UPDATE assessments
SET title = $2, description = $3, time_limit_seconds = $4, passing_score = $5
WHERE id = $1
RETURNING *;

-- name: DeleteAssessment :exec
DELETE FROM assessments WHERE id = $1;

-- ───────────────────────── Assessments ↔ Topics ─────────────────────────

-- name: AddTopicToAssessment :exec
INSERT INTO assessment_topics (assessment_id, topic_id)
VALUES ($1, $2)
ON CONFLICT (assessment_id, topic_id) DO NOTHING;

-- name: RemoveTopicFromAssessment :exec
DELETE FROM assessment_topics WHERE assessment_id = $1 AND topic_id = $2;

-- name: ListTopicsByAssessment :many
SELECT t.* FROM topics t
JOIN assessment_topics ato ON ato.topic_id = t.id
WHERE ato.assessment_id = $1
ORDER BY t.name ASC;

-- ───────────────────────── Assessments ↔ Questions ─────────────────────────

-- name: AddQuestionToAssessment :exec
INSERT INTO assessment_questions (assessment_id, question_id, position)
VALUES ($1, $2, $3)
ON CONFLICT (assessment_id, question_id) DO NOTHING;

-- name: RemoveQuestionFromAssessment :exec
DELETE FROM assessment_questions WHERE assessment_id = $1 AND question_id = $2;

-- name: UpdateAssessmentQuestionPosition :exec
UPDATE assessment_questions
SET position = $3
WHERE assessment_id = $1 AND question_id = $2;

-- name: ListQuestionsByAssessment :many
SELECT q.* FROM questions q
JOIN assessment_questions aq ON aq.question_id = q.id
WHERE aq.assessment_id = $1
ORDER BY aq.position ASC;
