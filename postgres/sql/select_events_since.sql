SELECT
    id,
    actor_id,
    created_on,
    content,
    version
FROM {{.ActorName}}_events
WHERE
    actor_id = $1 AND
    id > $2
ORDER BY id ASC;