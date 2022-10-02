SELECT
    parent_type,
    parent_id,
    child_type,
    child_id
FROM {{.ActorName}}_links
WHERE
    parent_type = $1
    AND parent_id = $2
ORDER BY id DESC;