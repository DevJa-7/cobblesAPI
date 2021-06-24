create table user_device_tokens (
	id bigserial not null
		constraint user_device_tokens_pkey
			primary key,
	user_id bigint,
	device_token text,
	enabled boolean,
	updated_at timestamp with time zone,
	created_at timestamp with time zone,
	endpoint_arn text
);