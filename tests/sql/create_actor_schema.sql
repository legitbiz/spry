CREATE IF NOT EXISTS TABLE {{.ActorName}}_commands (
    id              uuid            PRIMARY KEY,
    actor_id        uuid            NOT NULL,
    content         jsonb,
    created_on      timestamp with time zone            DEFAULT now(),
    vector          varchar(9192),
    version         bigint          NOT NULL,
)

CREATE INDEX IF NOT EXISTS {{.ActorName}}_command_actor_idx on {{.ActorName}}_commands(actor_id);

CREATE TABLE IF NOT EXISTS {{.ActorName}}_events (
    id              uuid            PRIMARY KEY,
    actor_id        uuid            NOT NULL,
    content         jsonb,
    created_on      timestamp with time zone            DEFAULT now(),
    vector          varchar(9192),
    version         bigint          NOT NULL,
);

CREATE INDEX IF NOT EXISTS {{.ActorName}}_event_actor_idx on {{.ActorName}}_events(actor_id);

CREATE TABLE IF NOT EXISTS {{.ActorName}}_id_map (
    id                      uuid        PRIMARY KEY,
    identifiers             jsonb       NOT NULL,
    actor_id                uuid        NOT NULL,
    starting_on             timestamp with time zone 	DEFAULT now(),
    UNIQUE(identifiers, actor_id)
);

CREATE INDEX IF NOT EXISTS {{.ActorName}}_id_map_actor_idx on {{.ActorName}}_id_map(actor_id);
CREATE INDEX IF NOT EXISTS {{.ActorName}}_id_map_ids_idx on {{.ActorName}}_id_map(identifiers);

CREATE TABLE IF NOT EXISTS {{.ActorName}}_snapshots (
	id								uuid	        PRIMARY KEY,
    actor_id                        uuid            NOT NULL,
	content							jsonb 			NOT NULL,
	last_command_id					uuid  	        NOT NULL,
	last_command_handled_on			timestamp with time zone  		NOT NULL,
	last_event_id					uuid 	        NOT NULL,
    last_event_applied_on			timestamp with time zone  		NOT NULL,
	vector							varchar(9192),
	version							bigint 			NOT NULL,
);

CREATE INDEX IF NOT EXISTS {{.ActorName}}_snapshot_actor_idx on {{.ActorName}}_snapshots(actor_id);