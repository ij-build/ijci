begin;

alter table projects drop column last_build_id;

drop table build_logs;
drop table builds;
drop table projects;
drop type build_status;

commit;
