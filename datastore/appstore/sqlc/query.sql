-- name: FindAppByID :one
SELECT * FROM app
WHERE app_id = $1 LIMIT 1;

-- name: FindAppByExternalID :one
SELECT * FROM app
WHERE app_extl_id = $1 LIMIT 1;

-- name: FindAppByName :one
SELECT a.*
FROM app a inner join org o on o.org_id = a.org_id
WHERE o.org_id = $1
  AND a.app_name = $2;

-- name: FindApps :many
SELECT * FROM app
ORDER BY app_name;

-- name: CreateApp :execresult
INSERT INTO app (app_id, org_id, app_extl_id, app_name, app_description, active, create_app_id, create_user_id,
                 create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: DeleteApp :exec
DELETE FROM app
WHERE app_id = $1;

-- name: FindAPIKeysByAppID :many
SELECT * FROM app_api_key
WHERE app_id = $1;

-- name: CreateAppAPIKey :execresult
INSERT INTO app_api_key (api_key, app_id, deactv_date, create_app_id, create_user_id,
                         create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: FindAppAPIKeysByAppExtlID :many
select a.app_id,
       a.app_extl_id,
       a.app_name,
       a.app_description,
       o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       aak.api_key,
       aak.deactv_date
from app a
         inner join org o on o.org_id = a.org_id
         inner join app_api_key aak on a.app_id = aak.app_id
where a.app_extl_id = $1;