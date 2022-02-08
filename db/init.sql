CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table users (
    id uuid DEFAULT uuid_generate_v1() PRIMARY KEY,
    username varchar(250) not null,
    first_name varchar(250) not null,
    last_name varchar(250) not null,
    middle_name varchar(250),
    password varchar(250) not null
);
create unique index uniq_users_id_idx on users (id);
create unique index uniq_users_name_idx on users (username);

create table tokens (
    user_id uuid not null,
    token varchar(1000) not null
);

create unique index uniq_tokens_user_id_idx on tokens (user_id);

