-- +migrate Up

-- +migrate StatementBegin
create function update_last_build(uuid, uuid) returns void as $$
begin
    update projects p
    set
        last_build_id = b.build_id,
        last_build_status = b.build_status,
        last_build_completed_at = b.completed_at
    from (
        select * from (
            select
                build_id,
                build_status,
                created_at,
                completed_at
            from builds
            where project_id = $1 and completed_at is not null and build_id is distinct from $2
        ) t
        union select null, null, null, null
        order by created_at desc nulls last
        limit 1
    ) b
    where p.project_id = $1;
end;
$$ language plpgsql;
-- +migrate StatementEnd

-- +migrate Down
drop function update_last_build(uuid, uuid);
