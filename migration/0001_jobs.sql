CREATE TABLE IF NOT EXISTS jobs (
		id uuid primary key,
		title text not null,
		tag text unique not null,
		expression varchar(25) not null,
		is_work boolean not null
	);

