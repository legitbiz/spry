INSERT INTO {{.ActorName}}_events (
    id,
    actor_id,
    content,
    created_on,
    version
) VALUES (
    $1, $2, $3, $4, $5
);