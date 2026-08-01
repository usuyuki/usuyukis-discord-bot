CREATE TABLE guild_channel_create_settings (
    id SERIAL PRIMARY KEY,
    guild_id TEXT NOT NULL UNIQUE,
    required_approvals INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE channel_create_proposals (
    id SERIAL PRIMARY KEY,
    guild_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    channel_name TEXT NOT NULL,
    proposer_id TEXT NOT NULL,
    resolved BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (channel_id, message_id)
);
