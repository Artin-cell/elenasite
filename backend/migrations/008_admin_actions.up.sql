-- Журнал действий администратора: кто, когда и что изменил в админке.
-- Пишется после успешного действия; сбой записи в журнал не ломает само действие.
-- Внешнего ключа на сущность нет: строка журнала переживает удаление записи или услуги.
-- admin_username копируется, чтобы было видно автора и после удаления администратора.
create table if not exists admin_actions (
    id             uuid primary key default uuid_generate_v4(),
    admin_id       uuid references admins (id) on delete set null,
    admin_username text not null default '',
    action         text not null,
    entity         text not null,
    entity_id      uuid,
    details        jsonb not null default '{}',
    created_at     timestamptz not null default now()
);
create index if not exists admin_actions_created_idx on admin_actions (created_at desc);
create index if not exists admin_actions_entity_idx on admin_actions (entity, entity_id);
