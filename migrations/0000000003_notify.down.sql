begin;

drop trigger build_logs_notify_event on build_logs;
drop trigger builds_notify_event on builds;
drop trigger projects_notify_event on projects;
drop function notify_event();

commit;
