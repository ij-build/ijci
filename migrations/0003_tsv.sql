-- +migrate Up
alter table projects add column tsv tsvector;
create index projects_tsv_idx on projects using gin(tsv);
alter table builds add column tsv tsvector;
create index builds_tsv_idx on builds using gin(tsv);

-- +migrate StatementBegin
create function projects_search_trigger() returns trigger as $$
begin
    new.tsv :=
        setweight(to_tsvector(coalesce(new.name, '')), 'A') ||
        setweight(to_tsvector(coalesce(new.repository_url, '')), 'D');
    return new;
end
$$ language plpgsql;
-- +migrate StatementEnd

create trigger update_projects_tsvector before insert or update
on projects for each row execute procedure projects_search_trigger();

-- +migrate StatementBegin
create function builds_search_trigger() returns trigger as $$
declare
    project projects%ROWTYPE;
begin
    select * into project from projects where project_id = new.project_id;

    new.tsv :=
        setweight(to_tsvector(coalesce(project.name, '')), 'A') ||
        setweight(to_tsvector(coalesce(project.repository_url, '')), 'D') ||
        setweight(to_tsvector(coalesce(new.commit_branch, '')), 'D') ||
        setweight(to_tsvector(coalesce(new.commit_message, '')), 'D') ||
        setweight(to_tsvector(coalesce(new.commit_hash, '')), 'D') ||
        setweight(to_tsvector(coalesce(new.commit_committer_name, '')), 'D') ||
        setweight(to_tsvector(coalesce(new.commit_committer_email, '')), 'D') ||
        setweight(to_tsvector(coalesce(new.commit_author_name, '')), 'D') ||
        setweight(to_tsvector(coalesce(new.commit_author_email, '')), 'D');
    return new;
end
$$ language plpgsql;
-- +migrate StatementEnd

create trigger update_builds_tsvector before insert or update
on builds for each row execute procedure builds_search_trigger();

-- +migrate Down
drop trigger update_projects_tsvector on projects;
drop trigger update_builds_tsvector on builds;
drop function projects_search_trigger();
drop function builds_search_trigger();
drop index projects_tsv_idx;
drop index builds_tsv_idx;
alter table projects drop column tsv;
alter table builds drop column tsv;
