CREATE TABLE keywords (
    id SERIAL PRIMARY KEY,
    guild_id TEXT NOT NULL,
    keyword TEXT NOT NULL,
    response TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (guild_id, keyword)
);

CREATE TABLE guild_notify_channels (
    id SERIAL PRIMARY KEY,
    guild_id TEXT NOT NULL,
    purpose TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (guild_id, purpose)
);
