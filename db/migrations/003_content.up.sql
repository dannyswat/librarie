-- Subjects
CREATE TABLE subjects (
    id              uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name            text        NOT NULL,
    description     text        NOT NULL DEFAULT '',
    cover_image_key text,
    position        integer     NOT NULL DEFAULT 0,
    created_by      uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at      timestamptz NOT NULL DEFAULT NOW()
);

-- Teacher ↔ Subject assignments (many-to-many)
CREATE TABLE teachers_subjects (
    teacher_id  uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id  uuid        NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    assigned_by uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    assigned_at timestamptz NOT NULL DEFAULT NOW(),
    PRIMARY KEY (teacher_id, subject_id)
);

-- Topics (shared across subjects)
CREATE TABLE topics (
    id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name        text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    created_by  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  timestamptz NOT NULL DEFAULT NOW()
);

-- Subject ↔ Topic assignments (many-to-many, ordered per subject)
CREATE TABLE subject_topics (
    subject_id uuid    NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    topic_id   uuid    NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    position   integer NOT NULL DEFAULT 0,
    PRIMARY KEY (subject_id, topic_id)
);

-- Contents (belong to one subject)
CREATE TABLE contents (
    id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    subject_id  uuid        NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    title       text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    position    integer     NOT NULL DEFAULT 0,
    created_by  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  timestamptz NOT NULL DEFAULT NOW(),
    updated_at  timestamptz NOT NULL DEFAULT NOW()
);

-- Content ↔ Topic tagging (many-to-many)
CREATE TABLE content_topics (
    content_id uuid NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    topic_id   uuid NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    PRIMARY KEY (content_id, topic_id)
);

-- Pages (ordered within a content item)
CREATE TABLE pages (
    id         uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    content_id uuid        NOT NULL REFERENCES contents(id) ON DELETE CASCADE,
    name       text        NOT NULL,
    position   integer     NOT NULL DEFAULT 0,
    created_by uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

-- Blocks (ordered within a page, type-specific JSONB payload)
CREATE TABLE blocks (
    id         uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    page_id    uuid        NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    type       text        NOT NULL,
    position   integer     NOT NULL DEFAULT 0,
    data       jsonb       NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);
