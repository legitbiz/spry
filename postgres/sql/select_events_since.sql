SELECT
    id,
    actor_id,
    created_on,
    content,
    version
FROM {{.ActorName}}_events
WHERE
    actor_id = $2 AND
    id > $1
ORDER BY id ASC;