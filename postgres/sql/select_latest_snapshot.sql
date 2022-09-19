SELECT
    actor_id,
    content,
    last_command_id,
    last_command_handled_on,
    last_event_id,
    last_event_applied_on,
    version
FROM {{.ActorName}}_snapshots
WHERE
    actor_id = $1
ORDER BY id DESC
LIMIT 1;