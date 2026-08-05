-- migrate:up
create table user_tokens (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users(id) on delete cascade,
    access_token text not null,
    refresh_token text,
    token_expiry timestamptz,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

create unique index idx_user_tokens_user_id on user_tokens(user_id);

-- migrate:down
drop table if exists user_tokens;
