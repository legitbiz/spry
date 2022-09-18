SELECT
    id,
    identifiers,
    actor_id,
    starting_on
FROM {{.ActorName}}_id_map
WHERE
    identifiers = $1
ORDER BY id DESC
LIMIT 1;