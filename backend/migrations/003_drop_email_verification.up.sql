drop table if exists email_verifications;

alter table clients drop column if exists email_verified;
