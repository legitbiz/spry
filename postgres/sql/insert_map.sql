INSERT INTO {{.ActorName}}_events (
    id
    identifiers
    actor_id
    starting_on
) VALUES (
    $1, $2, $3, $4
);