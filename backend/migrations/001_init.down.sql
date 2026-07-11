drop trigger if exists reviews_updated_at    on reviews;
drop trigger if exists news_updated_at       on news;
drop trigger if exists appointments_updated_at on appointments;
drop trigger if exists services_updated_at   on services;
drop trigger if exists clients_updated_at    on clients;
drop function if exists set_updated_at();

drop table if exists reviews;
drop table if exists news;
drop table if exists admins;
drop table if exists cancellation_tokens;
drop table if exists appointments;
drop table if exists services;
drop table if exists email_verifications;
drop table if exists clients;

drop extension if exists btree_gist;
drop extension if exists "uuid-ossp";
