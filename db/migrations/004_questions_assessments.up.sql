-- Questions (subject-scoped question bank)
CREATE TABLE questions (
    id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    subject_id  uuid        NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    stem_type   text        NOT NULL,
    stem_data   jsonb       NOT NULL DEFAULT '{}',
    type        text        NOT NULL,
    answer_data jsonb       NOT NULL DEFAULT '{}',
    tags        text[]      NOT NULL DEFAULT '{}',
    created_by  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  timestamptz NOT NULL DEFAULT NOW()
);

-- Question ↔ Topic tagging (many-to-many)
CREATE TABLE question_topics (
    question_id uuid NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    topic_id    uuid NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    PRIMARY KEY (question_id, topic_id)
);

-- Assessments
CREATE TABLE assessments (
    id                 uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    title              text        NOT NULL,
    description        text        NOT NULL DEFAULT '',
    subject_id         uuid        NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    time_limit_seconds integer,
    passing_score      integer,
    created_by         uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at         timestamptz NOT NULL DEFAULT NOW()
);

-- Assessment ↔ Topic tagging (many-to-many)
CREATE TABLE assessment_topics (
    assessment_id uuid NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    topic_id      uuid NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    PRIMARY KEY (assessment_id, topic_id)
);

-- Assessment ↔ Question (ordered list, questions reusable across assessments)
CREATE TABLE assessment_questions (
    assessment_id uuid    NOT NULL REFERENCES assessments(id) ON DELETE CASCADE,
    question_id   uuid    NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    position      integer NOT NULL DEFAULT 0,
    PRIMARY KEY (assessment_id, question_id)
);
