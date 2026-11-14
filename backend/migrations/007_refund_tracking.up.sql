-- Колонки для отслеживания возвратов в ЮKassa: ID возврата и его статус (pending/succeeded/failed).
-- Пока ЮKassa обрабатывает возврат (refund_status = 'pending'), админка показывает плашку «Возврат оформляется».
-- Вебхук refund.succeeded меняет статус на 'succeeded' и переводит payment_status в 'refunded'.
alter table appointments add column if not exists refund_id     text not null default '';
alter table appointments add column if not exists refund_status text not null default '';
