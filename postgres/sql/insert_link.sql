INSERT INTO {{.ActorName}}_link (
    id,
    parent_id,
    child_id
) VALUES (
    $1, $2, $3, $4
    )
ON CONFLICT DO NOTHING;