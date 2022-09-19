CREATE TABLE IF NOT EXISTS id_sort_test (
  id 	    	uuid PRIMARY KEY,
  ordering      bigint
);

CREATE TABLE IF NOT EXISTS player_commands (
    id              uuid            PRIMARY KEY,
    actor_id        uuid            NOT NULL,
    content         jsonb,
    created_on      timestamp with time zone            DEFAULT now(),
    vector          varchar(9192),
    version         bigint          NOT NULL
);

CREATE INDEX IF NOT EXISTS player_command_actor_idx on player_commands(actor_id);

CREATE TABLE IF NOT EXISTS player_events (
    id              uuid            PRIMARY KEY,
    actor_id        uuid            NOT NULL,
    content         jsonb,
    created_on      timestamp with time zone            DEFAULT now(),
    vector          varchar(9192),
    version         bigint          NOT NULL
);

CREATE INDEX IF NOT EXISTS player_event_actor_idx on player_events(actor_id);

CREATE TABLE IF NOT EXISTS player_id_map (
    id                      uuid        PRIMARY KEY,
    identifiers             jsonb       NOT NULL,
    actor_id                uuid        NOT NULL,
    starting_on             timestamp with time zone 	DEFAULT now(),
    UNIQUE(identifiers, actor_id)
);

CREATE INDEX IF NOT EXISTS player_id_map_actor_idx on player_id_map(actor_id);
CREATE INDEX IF NOT EXISTS player_id_map_ids_idx on player_id_map(identifiers);

CREATE TABLE IF NOT EXISTS player_snapshots (
	id								uuid	        PRIMARY KEY,
    actor_id                        uuid            NOT NULL,
	content							jsonb 			NOT NULL,
	last_command_id					uuid  	        NOT NULL,
	last_command_handled_on			timestamp with time zone  		NOT NULL,
	last_event_id					uuid 	        NOT NULL,
    last_event_applied_on			timestamp with time zone  		NOT NULL,
	vector							varchar(9192),
	version							bigint 			NOT NULL
);

CREATE INDEX IF NOT EXISTS player_snapshot_actor_idx on player_snapshots(actor_id);