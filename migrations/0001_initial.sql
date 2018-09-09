-- +migrate Up
create type build_status as enum (
    'queued',
    'in-progress',
    'complete'
);

create table builds (
    build_id uuid primary key,
    repository_url text not null,
    build_status build_status not null,
	created_at timestamp with time zone not null,
	updated_at timestamp with time zone not null
);

-- +migrate Down
drop table builds;
drop type build_status;
