-- name: FindAppByID :one
-- FindAppByID selects an app given its app_id.
SELECT a.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       a.app_id,
       a.app_extl_id,
       a.app_name,
       a.app_description
FROM app a
         INNER JOIN org o on o.org_id = a.org_id
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
WHERE a.app_id = $1;

-- name: FindAppByIDWithAudit :one
-- FindAppByID selects an app given its app_id.
-- FindAppByID also includes audit information as part of the return.
SELECT a.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       a.app_id,
       a.app_extl_id,
       a.app_name,
       a.app_description,
       a.create_app_id,
       ca.org_id          create_app_org_id,
       ca.app_extl_id     create_app_extl_id,
       ca.app_name        create_app_name,
       ca.app_description create_app_description,
       a.create_user_id,
       cu.first_name     create_user_first_name,
       cu.last_name      create_user_last_name,
       a.create_timestamp,
       a.update_app_id,
       ua.org_id          update_app_org_id,
       ua.app_extl_id     update_app_extl_id,
       ua.app_name        update_app_name,
       ua.app_description update_app_description,
       a.update_user_id,
       uu.first_name     update_user_first_name,
       uu.last_name      update_user_last_name,
       a.update_timestamp
FROM app a
         INNER JOIN org o on o.org_id = a.org_id
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
         INNER JOIN app ca on ca.app_id = a.create_app_id
         INNER JOIN app ua on ua.app_id = a.update_app_id
         LEFT JOIN users cu on cu.user_id = a.create_user_id
         LEFT JOIN users uu on uu.user_id = a.update_user_id
WHERE a.app_id = $1;

-- name: FindAppByExternalID :one
-- FindAppByExternalID selects an app given its app_extl_id.
SELECT a.app_id,
       a.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       a.app_extl_id,
       a.app_name,
       a.app_description
FROM app a
         INNER JOIN org o on o.org_id = a.org_id
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
WHERE a.app_extl_id = $1;

-- name: FindAppByExternalIDWithAudit :one
-- FindAppByExternalID selects an app given its app_extl_id.
-- FindAppByExternalID also includes audit information as part of the return.
SELECT a.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       a.app_id,
       a.app_extl_id,
       a.app_name,
       a.app_description,
       a.create_app_id,
       ca.org_id          create_app_org_id,
       ca.app_extl_id     create_app_extl_id,
       ca.app_name        create_app_name,
       ca.app_description create_app_description,
       a.create_user_id,
       cu.first_name     create_user_first_name,
       cu.last_name      create_user_last_name,
       a.create_timestamp,
       a.update_app_id,
       ua.org_id          update_app_org_id,
       ua.app_extl_id     update_app_extl_id,
       ua.app_name        update_app_name,
       ua.app_description update_app_description,
       a.update_user_id,
       uu.first_name     update_user_first_name,
       uu.last_name      update_user_last_name,
       a.update_timestamp
FROM app a
         INNER JOIN org o on o.org_id = a.org_id
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
         INNER JOIN app ca on ca.app_id = a.create_app_id
         INNER JOIN app ua on ua.app_id = a.update_app_id
         LEFT JOIN users cu on cu.user_id = a.create_user_id
         LEFT JOIN users uu on uu.user_id = a.update_user_id
WHERE a.app_extl_id = $1;

-- name: FindAppByName :one
-- FindAppByName selects an app given its app_name.
SELECT a.app_id,
       a.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       a.app_extl_id,
       a.app_name,
       a.app_description
FROM app a
         INNER JOIN org o on o.org_id = a.org_id
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
WHERE o.org_id = $1
  AND a.app_name = $2;

-- name: FindAppByProviderClientID :one
-- FindAppByProviderClientID selects an app given an external auth provider client ID.
SELECT a.app_id,
       a.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       a.app_extl_id,
       a.app_name,
       a.app_description
FROM app a
         INNER JOIN org o on o.org_id = a.org_id
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
WHERE a.auth_provider_client_id = $1;

-- name: FindApps :many
-- FindApps returns every app.
SELECT * FROM app
ORDER BY app_name;

-- name: FindAppsByOrg :many
-- FindAppsByOrg returns all apps for a given Organization.
SELECT * FROM app
WHERE org_id = $1;

-- name: FindAppsWithAudit :many
-- FindAppsWithAudit returns every app.
-- FindAppsWithAudit also includes audit information as part of the return.
SELECT a.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       a.app_id,
       a.app_extl_id,
       a.app_name,
       a.app_description,
       a.create_app_id,
       ca.org_id          create_app_org_id,
       ca.app_extl_id     create_app_extl_id,
       ca.app_name        create_app_name,
       ca.app_description create_app_description,
       a.create_user_id,
       cu.first_name     create_user_first_name,
       cu.last_name      create_user_last_name,
       a.create_timestamp,
       a.update_app_id,
       ua.org_id          update_app_org_id,
       ua.app_extl_id     update_app_extl_id,
       ua.app_name        update_app_name,
       ua.app_description update_app_description,
       a.update_user_id,
       uu.first_name     update_user_first_name,
       uu.last_name      update_user_last_name,
       a.update_timestamp
FROM app a
         INNER JOIN org o on o.org_id = a.org_id
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
         INNER JOIN app ca on ca.app_id = a.create_app_id
         INNER JOIN app ua on ua.app_id = a.update_app_id
         LEFT JOIN users cu on cu.user_id = a.create_user_id
         LEFT JOIN users uu on uu.user_id = a.update_user_id;

-- name: CreateApp :execrows
-- CreateApp inserts a new app into the app table.
INSERT INTO app (app_id, org_id, app_extl_id, app_name, app_description,
                 auth_provider_id, auth_provider_client_id,
                 create_app_id, create_user_id, create_timestamp,
                 update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: UpdateApp :execrows
-- UpdateApp updates an app given its app_id.
UPDATE app
SET app_name        = $1,
    app_description = $2,
    update_app_id    = $3,
    update_user_id   = $4,
    update_timestamp = $5
WHERE app_id = $6;


-- name: DeleteApp :execrows
-- DeleteApp deletes an app given its app_id.
DELETE FROM app
WHERE app_id = $1;

-- name: DeleteAppAPIKey :execrows
-- DeleteAppAPIKey deletes an app API key given the key.
DELETE FROM app_api_key
WHERE api_key = $1;

-- name: DeleteAppAPIKeys :execrows
-- DeleteAppAPIKeys deletes all API keys for a given app_id.
DELETE FROM app_api_key
WHERE app_id = $1;

-- name: FindAPIKeysByAppID :many
-- FindAPIKeysByAppID selects all API keys for a given app_id.
SELECT * FROM app_api_key
WHERE app_id = $1;

-- name: CreateAppAPIKey :execrows
-- CreateAppAPIKey inserts an app API key into the app_api_key table.
INSERT INTO app_api_key (api_key, app_id, deactv_date, create_app_id, create_user_id,
                         create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: FindAppAPIKeysByAppExtlID :many
-- FindAppAPIKeysByAppExtlID selects all app API keys given an app external ID.
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