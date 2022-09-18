INSERT INTO {{.ActorName}}_events (
    id,
    actor_id,
    content,
    last_command_id,
    last_command_handled_on,
    last_event_id,
    last_event_applied_on,
    version
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);