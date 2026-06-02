CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Users
CREATE TABLE users (
    id            uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    username      text        NOT NULL UNIQUE,
    email         text        NOT NULL UNIQUE,
    password_hash text        NOT NULL,
    role          text        NOT NULL CHECK (role IN ('admin', 'teacher', 'student')),
    created_at    timestamptz NOT NULL DEFAULT NOW()
);

-- Passkeys (WebAuthn / FIDO2 credentials)
CREATE TABLE passkeys (
    id            uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id       uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id bytea       NOT NULL UNIQUE,
    public_key    bytea       NOT NULL,
    sign_count    bigint      NOT NULL DEFAULT 0,
    created_at    timestamptz NOT NULL DEFAULT NOW()
);

-- Sessions
CREATE TABLE sessions (
    id           uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id      uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   text        NOT NULL UNIQUE,
    expires_at   timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL DEFAULT NOW(),
    created_at   timestamptz NOT NULL DEFAULT NOW()
);

-- Login attempt audit log
CREATE TABLE login_attempts (
    id                   uuid        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    username             text        NOT NULL,
    ip                   text        NOT NULL,
    password_fingerprint text,
    attempted_at         timestamptz NOT NULL DEFAULT NOW(),
    success              boolean     NOT NULL DEFAULT FALSE
);

-- Token-bucket rate limiter state
CREATE TABLE login_rate_limits (
    scope_type text        NOT NULL,
    scope_key  text        NOT NULL,
    tokens     integer     NOT NULL DEFAULT 5,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    PRIMARY KEY (scope_type, scope_key)
);
