create table urls(
    url_key text primary key,
    original_url text not null constraint urls_original_url_unique unique
);
