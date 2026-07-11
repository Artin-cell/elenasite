create extension if not exists "uuid-ossp";

create table clients (
    id              uuid primary key default uuid_generate_v4(),
    first_name      text not null,
    last_name       text not null,
    patronym        text not null default '',
    phone           text not null,
    email           text not null,
    email_verified  boolean not null default false,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now()
);

create unique index clients_email_idx on clients (email);
create unique index clients_phone_idx on clients (phone);

create table email_verifications (
    id          uuid primary key default uuid_generate_v4(),
    client_id   uuid not null references clients (id) on delete cascade,
    code        char(6) not null,
    expires_at  timestamptz not null,
    used_at     timestamptz,
    created_at  timestamptz not null default now()
);

create index ev_client_idx on email_verifications (client_id);

create table services (
    id            uuid primary key default uuid_generate_v4(),
    title         text not null,
    description   text not null default '',
    format        text not null check (format in ('online', 'offline', 'both')),
    price_kopeks  bigint not null default 0 check (price_kopeks >= 0),
    duration_min  int not null default 60 check (duration_min > 0),
    is_demo       boolean not null default false,
    is_active     boolean not null default true,
    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now()
);

create table appointments (
    id              uuid primary key default uuid_generate_v4(),
    client_id       uuid not null references clients (id) on delete restrict,
    service_id      uuid not null references services (id) on delete restrict,
    format          text not null check (format in ('online', 'offline')),
    starts_at       timestamptz not null,
    ends_at         timestamptz not null,
    status          text not null default 'pending'
                        check (status in ('pending','confirmed','cancelled','completed','no_show')),
    notes           text not null default '',
    payment_status  text not null default 'stub',
    payment_id      text not null default '',
    amount_kopeks   bigint not null default 0,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now(),

    constraint ends_after_starts check (ends_at > starts_at)
);

create index appt_client_idx  on appointments (client_id);
create index appt_service_idx on appointments (service_id);
create index appt_starts_idx  on appointments (starts_at);
create index appt_status_idx  on appointments (status);

create extension if not exists btree_gist;

alter table appointments
    add constraint no_overlapping_appointments
    exclude using gist (
        tstzrange(starts_at, ends_at) with &&
    )
    where (status not in ('cancelled'));

create table cancellation_tokens (
    id              uuid primary key default uuid_generate_v4(),
    appointment_id  uuid not null references appointments (id) on delete cascade,
    token           text not null unique,
    expires_at      timestamptz not null,
    used_at         timestamptz,
    created_at      timestamptz not null default now()
);

create index ct_token_idx on cancellation_tokens (token);

create table admins (
    id            uuid primary key default uuid_generate_v4(),
    username      text not null unique,
    password_hash text not null,
    created_at    timestamptz not null default now()
);

create table news (
    id            uuid primary key default uuid_generate_v4(),
    title         text not null,
    body          text not null default '',
    published_at  timestamptz,
    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now()
);

create table reviews (
    id           uuid primary key default uuid_generate_v4(),
    author_name  text not null,
    body         text not null,
    is_visible   boolean not null default false,
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now()
);

create or replace function set_updated_at()
returns trigger language plpgsql as $$
begin
    new.updated_at = now();
    return new;
end;
$$;

create trigger clients_updated_at    before update on clients    for each row execute function set_updated_at();
create trigger services_updated_at   before update on services   for each row execute function set_updated_at();
create trigger appointments_updated_at before update on appointments for each row execute function set_updated_at();
create trigger news_updated_at       before update on news       for each row execute function set_updated_at();
create trigger reviews_updated_at    before update on reviews    for each row execute function set_updated_at();
