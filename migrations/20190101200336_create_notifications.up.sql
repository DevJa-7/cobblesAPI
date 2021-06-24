create table notifications (
	id bigserial not null
		constraint notifications_pkey
			primary key,
	user_id bigint,
	data jsonb,
    read boolean default false,
	created_at timestamp with time zone,
	updated_at timestamp with time zone
);