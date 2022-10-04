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

CREATE TABLE IF NOT EXISTS motorist_commands (
    id              uuid            PRIMARY KEY,
    actor_id        uuid            NOT NULL,
    content         jsonb,
    created_on      timestamp with time zone            DEFAULT now(),
    vector          varchar(9192),
    version         bigint          NOT NULL
);

CREATE INDEX IF NOT EXISTS motorist_command_actor_idx on motorist_commands(actor_id);

CREATE TABLE IF NOT EXISTS motorist_events (
    id              uuid            PRIMARY KEY,
    actor_id        uuid            NOT NULL,
    content         jsonb,
    created_on      timestamp with time zone            DEFAULT now(),
    vector          varchar(9192),
    version         bigint          NOT NULL
);

CREATE INDEX IF NOT EXISTS motorist_event_actor_idx on motorist_events(actor_id);

CREATE TABLE IF NOT EXISTS motorist_id_map (
    id                      uuid        PRIMARY KEY,
    identifiers             jsonb       NOT NULL,
    actor_id                uuid        NOT NULL,
    starting_on             timestamp with time zone 	DEFAULT now(),
    UNIQUE(identifiers, actor_id)
);

CREATE INDEX IF NOT EXISTS motorist_id_map_actor_idx on motorist_id_map(actor_id);
CREATE INDEX IF NOT EXISTS motorist_id_map_ids_idx on motorist_id_map(identifiers);

CREATE TABLE IF NOT EXISTS motorist_links (
    id                      uuid            PRIMARY KEY,
    parent_type             varchar(128)    NOT NULL,
    parent_id               uuid            NOT NULL,
    child_type              varchar(128)    NOT NULL,
    child_id                uuid            NOT NULL,
    active                  bool            DEFAULT(true),
    starting_on             timestamp with time zone 	DEFAULT now(),
    UNIQUE(parent_id, child_id)
);

CREATE INDEX IF NOT EXISTS motorist_link_parent_idx on motorist_links(parent_id);
CREATE INDEX IF NOT EXISTS motorist_link_child_idx on motorist_links(child_id);

CREATE TABLE IF NOT EXISTS motorist_snapshots (
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

CREATE INDEX IF NOT EXISTS motorist_snapshot_actor_idx on motorist_snapshots(actor_id);

CREATE TABLE IF NOT EXISTS vehicle_commands (
    id              uuid            PRIMARY KEY,
    actor_id        uuid            NOT NULL,
    content         jsonb,
    created_on      timestamp with time zone            DEFAULT now(),
    vector          varchar(9192),
    version         bigint          NOT NULL
);

CREATE INDEX IF NOT EXISTS vehicle_command_actor_idx on vehicle_commands(actor_id);

CREATE TABLE IF NOT EXISTS vehicle_events (
    id              uuid            PRIMARY KEY,
    actor_id        uuid            NOT NULL,
    content         jsonb,
    created_on      timestamp with time zone            DEFAULT now(),
    vector          varchar(9192),
    version         bigint          NOT NULL
);

CREATE INDEX IF NOT EXISTS vehicle_event_actor_idx on vehicle_events(actor_id);

CREATE TABLE IF NOT EXISTS vehicle_id_map (
    id                      uuid        PRIMARY KEY,
    identifiers             jsonb       NOT NULL,
    actor_id                uuid        NOT NULL,
    starting_on             timestamp with time zone 	DEFAULT now(),
    UNIQUE(identifiers, actor_id)
);

CREATE INDEX IF NOT EXISTS vehicle_id_map_actor_idx on vehicle_id_map(actor_id);
CREATE INDEX IF NOT EXISTS vehicle_id_map_ids_idx on vehicle_id_map(identifiers);

CREATE TABLE IF NOT EXISTS vehicle_links (
    id                      uuid            PRIMARY KEY,
    parent_type             varchar(128)    NOT NULL,
    parent_id               uuid            NOT NULL,
    child_type              varchar(128)    NOT NULL,
    child_id                uuid            NOT NULL,
    active                  bool            DEFAULT(true),
    starting_on             timestamp with time zone 	DEFAULT now(),
    UNIQUE(parent_id, child_id)
);

CREATE INDEX IF NOT EXISTS vehicle_link_parent_idx on vehicle_links(parent_id);
CREATE INDEX IF NOT EXISTS vehicle_link_child_idx on vehicle_links(child_id);

CREATE TABLE IF NOT EXISTS vehicle_snapshots (
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

CREATE INDEX IF NOT EXISTS vehicle_snapshot_actor_idx on vehicle_snapshots(actor_id);