-- MUDRO 012_password_auth.sql

alter table users add column if not exists password_hash text;
