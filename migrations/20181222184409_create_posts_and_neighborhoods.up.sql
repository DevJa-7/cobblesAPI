create table posts
(
	id bigserial not null
		constraint posts_pkey
			primary key,
	user_id bigint,
	title text,
	description text,
	preview_url text,
	updated_at timestamp with time zone,
	created_at timestamp with time zone,
	neighborhood_id bigint,
	kind text,
	processing boolean,
	media_kind text,
	uploaded_media_url text,
	media_url text,
	tags text[],
	media_metadata jsonb,
	poster text,
	media jsonb,
	preview jsonb,
	uploaded_media jsonb
);


create table neighborhoods
(
	id bigserial not null
		constraint neighborhoods_pkey
			primary key,
	name text,
	slug text,
	created_at timestamp with time zone,
	updated_at timestamp with time zone
);

insert into neighborhoods (name, slug, created_at, updated_at)
	values ('Southie', 'southie', now(), now());
