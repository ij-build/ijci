-- +migrate Up
create type build_status as enum (
    'queued',
    'in-progress',
    'errored',
    'failed',
    'succeeded'
);

create table builds (
    build_id uuid primary key,
    repository_url text not null,
    build_status build_status not null,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone not null
);

create table build_logs (
    build_log_id uuid primary key,
    build_id uuid not null references builds on delete cascade,
    name text not null,
    key text not null,
    uploaded_at timestamp with time zone not null
);

-- +migrate Down
drop table build_logs;
drop table builds;

drop type build_status;
