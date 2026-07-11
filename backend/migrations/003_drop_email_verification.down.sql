alter table clients add column email_verified boolean not null default false;

create table email_verifications (
    id          uuid primary key default uuid_generate_v4(),
    client_id   uuid not null references clients (id) on delete cascade,
    code        char(6) not null,
    expires_at  timestamptz not null,
    used_at     timestamptz,
    created_at  timestamptz not null default now()
);

create index ev_client_idx on email_verifications (client_id);
