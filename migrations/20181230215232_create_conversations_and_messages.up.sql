create table conversations (
    id bigserial primary key,

    post_id bigint references posts (id) not null,
    started_by_user_id bigint references users (id) not null,

    created_at timestamptz default now() not null,

    unique (started_by_user_id, post_id)
);
 
create table conversation_has_users (
    conversation_id bigint references conversations (id) not null,
    user_id bigint references users (id) not null,

    unique (conversation_id, user_id)
);

create table messages (
    id bigserial primary key,
    from_user_id bigint references users (id),
    conversation_id bigint references conversations (id),
    body text not null,

    created_at timestamptz default now() not null
);