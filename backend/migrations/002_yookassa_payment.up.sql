alter table appointments
    add column payment_mode text not null default 'full'
        check (payment_mode in ('full', 'prepay_50'));

alter table appointments
    drop constraint if exists appointments_payment_status_check;

alter table appointments
    add constraint appointments_payment_status_check
    check (payment_status in ('pending', 'paid', 'failed', 'refunded'));

alter table appointments
    alter column payment_status set default 'pending';

update appointments set payment_status = 'pending' where payment_status = 'stub';

create index appt_payment_id_idx on appointments (payment_id) where payment_id != '';
