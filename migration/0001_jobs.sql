-- +migrate Up

CREATE TYPE "CRONJOB_STATUS" AS ENUM (
	'created',
	'working',
	'suspended',
	'done',
	'failed',
    'cancelled'
);

CREATE TABLE IF NOT EXISTS jobs (
	tag uuid primary key,
	id uuid not null,
	expression varchar(25) not null,
    status "CRONJOB_STATUS" DEFAULT 'created',
    status_description text not null DEFAULT '',
    function_name varchar(50) not null,
    function_fields jsonb not null,
    limit int DEFAULT 1
);

ALTER TABLE jobs ADD CONSTRAINT unique_title_operation UNIQUE (id, function_name);

-- +migrate Down

DROP TYPE IF EXISTS "CRONJOB_OPERATION";

DROP TABLE IF EXISTS jobs;
