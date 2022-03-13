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
       a.org_id           create_app_org_id,
       a.app_extl_id      create_app_extl_id,
       a.app_name         create_app_name,
       a.app_description  create_app_description,
       m.create_user_id,
       ou.username        create_username,
       ou.org_id          create_user_org_id,
       pp.first_name      create_user_first_name,
       pp.last_name       create_user_last_name,
       m.create_timestamp,
       m.update_app_id,
       a2.org_id          update_app_org_id,
       a2.app_extl_id     update_app_extl_id,
       a2.app_name        update_app_name,
       a2.app_description update_app_description,
       m.update_user_id,
       ou2.username       update_username,
       ou2.org_id         update_user_org_id,
       pp2.first_name     update_user_first_name,
       pp2.last_name      update_user_last_name,
       m.update_timestamp
FROM movie m
         INNER JOIN app a on a.app_id = m.create_app_id
         INNER JOIN app a2 on a2.app_id = m.update_app_id
         LEFT JOIN org_user ou on ou.user_id = m.create_user_id
         INNER JOIN person_profile pp on pp.person_profile_id = ou.person_profile_id
         LEFT JOIN org_user ou2 on ou2.user_id = m.update_user_id
         INNER JOIN person_profile pp2 on pp2.person_profile_id = ou2.person_profile_id
WHERE m.extl_id = $1;

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
       a.org_id           create_app_org_id,
       a.app_extl_id      create_app_extl_id,
       a.app_name         create_app_name,
       a.app_description  create_app_description,
       m.create_user_id,
       ou.username        create_username,
       ou.org_id          create_user_org_id,
       pp.first_name      create_user_first_name,
       pp.last_name       create_user_last_name,
       m.create_timestamp,
       m.update_app_id,
       a2.org_id          update_app_org_id,
       a2.app_extl_id     update_app_extl_id,
       a2.app_name        update_app_name,
       a2.app_description update_app_description,
       m.update_user_id,
       ou2.username       update_username,
       ou2.org_id         update_user_org_id,
       pp2.first_name     update_user_first_name,
       pp2.last_name      update_user_last_name,
       m.update_timestamp
FROM movie m
         INNER JOIN app a on a.app_id = m.create_app_id
         INNER JOIN app a2 on a2.app_id = m.update_app_id
         LEFT JOIN org_user ou on ou.user_id = m.create_user_id
         INNER JOIN person_profile pp on pp.person_profile_id = ou.person_profile_id
         LEFT JOIN org_user ou2 on ou2.user_id = m.update_user_id
         INNER JOIN person_profile pp2 on pp2.person_profile_id = ou2.person_profile_id;

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
