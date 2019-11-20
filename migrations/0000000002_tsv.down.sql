begin;

drop trigger update_projects_tsvector on projects;
drop trigger update_builds_tsvector on builds;
drop function projects_search_trigger();
drop function builds_search_trigger();
drop index projects_tsv_idx;
drop index builds_tsv_idx;
alter table projects drop column tsv;
alter table builds drop column tsv;

commit;
