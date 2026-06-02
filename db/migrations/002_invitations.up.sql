CREATE TABLE invitations (
    id          uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    email       text        NOT NULL,
    role        text        NOT NULL CHECK (role IN ('admin', 'teacher', 'student')),
    token_hash  text        NOT NULL UNIQUE,
    invited_by  uuid        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    expires_at  timestamptz NOT NULL,
    accepted_at timestamptz,
    created_at  timestamptz NOT NULL DEFAULT NOW()
);
