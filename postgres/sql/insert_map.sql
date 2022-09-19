INSERT INTO {{.ActorName}}_id_map (
    id,
    identifiers,
    actor_id,
    starting_on
) VALUES (
    $1, $2, $3, $4
);