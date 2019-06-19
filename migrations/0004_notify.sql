-- +migrate Up

-- +migrate StatementBegin
create function notify_event() returns trigger as $$
declare
    data json;
    notification json;
begin
    if (tg_op = 'DELETE') then
        data = row_to_json(OLD);
    else
        data = row_to_json(new);
    end if;

    notification = json_build_object(
        'table', TG_TABLE_NAME,
        'action', TG_OP,
        'data', data
    );

    perform pg_notify('events', notification::text);
    return null;
end;
$$ language plpgsql;
-- +migrate StatementEnd

create trigger build_logs_notify_event after insert or update or delete on build_logs for each row execute procedure notify_event();
create trigger builds_notify_event after insert or update or delete on builds for each row execute procedure notify_event();
create trigger projects_notify_event after insert or update or delete on projects for each row execute procedure notify_event();

-- +migrate Down
drop trigger build_logs_notify_event on build_logs;
drop trigger builds_notify_event on builds;
drop trigger projects_notify_event on projects;
drop function notify_event();
