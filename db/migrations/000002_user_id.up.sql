create table app_parameters(
    id text primary key,
    value_str text not null
);

insert into app_parameters values(
    'jwt_signing_key', md5(random()::text)
);

alter table urls add column user_id text;
create index urls_user_id_idx on urls(user_id);
