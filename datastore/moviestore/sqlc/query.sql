-- name: CreateMovie :execresult
INSERT INTO movie (movie_id, extl_id, title, rated, released, run_time, director, writer,
                   create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: FindMovieByExternalID :one
SELECT * FROM movie
WHERE extl_id = $1;

-- name: FindMovies :many
SELECT * FROM movie;

-- name: UpdateMovie :exec
UPDATE movie
SET title            = $1,
    rated            = $2,
    released         = $3,
    run_time         = $4,
    director         = $5,
    writer           = $6,
    update_app_id    = $7,
    update_user_id   = $8,
    update_timestamp = $9
WHERE movie_id = $10;

-- name: DeleteMovie :exec
DELETE FROM movie
WHERE movie_id = $1;
