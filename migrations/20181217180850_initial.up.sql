create table if not exists login_codes (
	id bigserial not null
		constraint login_codes_pkey
			primary key,
	login_code text,
	phone_number text,
	expires_at timestamp with time zone,
	created_at timestamp with time zone,
	updated_at timestamp with time zone,
	enabled boolean
);

create table if not exists users (
	id bigserial not null
		constraint users_pkey
			primary key,
	name text,
	phone_number text,
	photo_url text,
	zip_code text,
	updated_at timestamp with time zone,
	created_at timestamp with time zone,
	cognito_username text
);

create unique index if not exists users_phone_number_uindex
	on users (phone_number);
