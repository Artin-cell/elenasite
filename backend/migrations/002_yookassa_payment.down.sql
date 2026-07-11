drop index if exists appt_payment_id_idx;

alter table appointments
    drop constraint if exists appointments_payment_status_check;

alter table appointments
    alter column payment_status set default 'stub';

alter table appointments
    add constraint appointments_payment_status_check
    check (payment_status in ('stub', 'pending', 'paid', 'refunded'));

alter table appointments
    drop column if exists payment_mode;
