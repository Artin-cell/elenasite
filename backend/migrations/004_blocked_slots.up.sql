create table blocked_slots (
    id           uuid primary key default uuid_generate_v4(),
    blocked_date date not null,
    slot_time    text,              -- 'HH:MM', null = блокирует весь день
    reason       text not null default '',
    created_at   timestamptz not null default now()
);

create index bs_date_idx on blocked_slots (blocked_date);
create unique index bs_full_day_idx on blocked_slots (blocked_date) where slot_time is null;
create unique index bs_slot_idx on blocked_slots (blocked_date, slot_time) where slot_time is not null;