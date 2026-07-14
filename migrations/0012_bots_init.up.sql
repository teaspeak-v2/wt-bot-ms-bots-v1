create extension if not exists pgcrypto;

create table if not exists bots (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    teamspeak_id uuid not null,
    owner_id uuid not null,
    nickname text not null,
    greeting text not null default 'Welcome to the server!',
    help_message text not null default 'Available commands: !help',
    enabled boolean not null default true,
    status text not null default 'offline',
    api_key text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index if not exists idx_bots_status on bots(status);
create index if not exists idx_bots_enabled on bots(enabled);
create index if not exists idx_bots_owner_id on bots(owner_id);
create index if not exists idx_bots_teamspeak_id on bots(teamspeak_id);
create index if not exists idx_bots_created_at on bots(created_at desc);

create or replace function set_updated_at()
returns trigger language plpgsql as $$
begin
    new.updated_at = now();
    return new;
end;
$$;

drop trigger if exists trg_bots_updated_at on bots;
create trigger trg_bots_updated_at before update on bots
for each row execute function set_updated_at();
