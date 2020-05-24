begin;

create function notify_event() returns trigger as $$
declare
    data json;
    notification json;
begin
    if (tg_op = 'DELETE') then
        data = to_jsonb(OLD) - 'content';
    else
        data = to_jsonb(NEW) - 'content';
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

create trigger build_logs_notify_event after insert or update or delete on build_logs for each row execute procedure notify_event();
create trigger builds_notify_event after insert or update or delete on builds for each row execute procedure notify_event();
create trigger projects_notify_event after insert or update or delete on projects for each row execute procedure notify_event();

commit;
