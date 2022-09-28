INSERT INTO {{.ActorName}}_link (
    id,
    parent_type,
    parent_id,
    child_type,
    child_id
) VALUES (
    $1, $2, $3, $4, $5
    )
ON CONFLICT DO NOTHING;