-- name: CreateMovie :execresult
INSERT INTO movie (movie_id, extl_id, title, rated, released, run_time, director, writer,
                   create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: FindMovieByExternalID :one
SELECT m.*
FROM movie m
WHERE m.extl_id = $1;

-- name: FindMovieByExternalIDWithAudit :one
SELECT m.movie_id,
       m.extl_id,
       m.title,
       m.rated,
       m.released,
       m.run_time,
       m.director,
       m.writer,
       m.create_app_id,
       ca.org_id          create_app_org_id,
       ca.app_extl_id     create_app_extl_id,
       ca.app_name        create_app_name,
       ca.app_description create_app_description,
       m.create_user_id,
       cu.first_name     create_user_first_name,
       cu.last_name      create_user_last_name,
       m.create_timestamp,
       m.update_app_id,
       ua.org_id          update_app_org_id,
       ua.app_extl_id     update_app_extl_id,
       ua.app_name        update_app_name,
       ua.app_description update_app_description,
       m.update_user_id,
       uu.first_name     update_user_first_name,
       uu.last_name      update_user_last_name,
       m.update_timestamp
FROM movie m
         INNER JOIN app ca on ca.app_id = m.create_app_id
         INNER JOIN app ua on ua.app_id = m.update_app_id
         LEFT JOIN users cu on cu.user_id = m.create_user_id
         LEFT JOIN users uu on uu.user_id = m.update_user_id
WHERE m.extl_id = $1;

-- name: FindMoviesByTitle :many
SELECT m.*
FROM movie m
WHERE m.title = $1;

-- name: FindMovies :many
SELECT m.movie_id,
       m.extl_id,
       m.title,
       m.rated,
       m.released,
       m.run_time,
       m.director,
       m.writer,
       m.create_app_id,
       ca.org_id          create_app_org_id,
       ca.app_extl_id     create_app_extl_id,
       ca.app_name        create_app_name,
       ca.app_description create_app_description,
       m.create_user_id,
       cu.first_name     create_user_first_name,
       cu.last_name      create_user_last_name,
       m.create_timestamp,
       m.update_app_id,
       ua.org_id          update_app_org_id,
       ua.app_extl_id     update_app_extl_id,
       ua.app_name        update_app_name,
       ua.app_description update_app_description,
       m.update_user_id,
       uu.first_name     update_user_first_name,
       uu.last_name      update_user_last_name,
       m.update_timestamp
FROM movie m
         INNER JOIN app ca on ca.app_id = m.create_app_id
         INNER JOIN app ua on ua.app_id = m.update_app_id
         LEFT JOIN users cu on cu.user_id = m.create_user_id
         LEFT JOIN users uu on uu.user_id = m.update_user_id;

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

-- name: DeleteMovie :execrows
DELETE FROM movie
WHERE movie_id = $1;
