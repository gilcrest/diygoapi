-- name: FindOrgByID :one
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       o.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
WHERE o.org_id = $1;

-- name: FindOrgByIDWithAudit :one
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       o.create_app_id,
       a.org_id           create_app_org_id,
       a.app_extl_id      create_app_extl_id,
       a.app_name         create_app_name,
       a.app_description  create_app_description,
       o.create_user_id,
       cu.first_name      create_user_first_name,
       cu.last_name       create_user_last_name,
       o.create_timestamp,
       o.update_app_id,
       a2.org_id          update_app_org_id,
       a2.app_extl_id     update_app_extl_id,
       a2.app_name        update_app_name,
       a2.app_description update_app_description,
       o.update_user_id,
       uu.first_name      update_user_first_name,
       uu.last_name       update_user_last_name,
       o.update_timestamp
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
         INNER JOIN app a on a.app_id = o.create_app_id
         INNER JOIN app a2 on a2.app_id = o.update_app_id
         INNER JOIN users cu on cu.user_id = o.create_user_id
         INNER JOIN users uu on uu.user_id = o.update_user_id
WHERE o.org_id = $1;

-- name: FindOrgByExtlID :one
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       o.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
WHERE org_extl_id = $1;

-- name: FindOrgByExtlIDWithAudit :one
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       o.create_app_id,
       a.org_id           create_app_org_id,
       a.app_extl_id      create_app_extl_id,
       a.app_name         create_app_name,
       a.app_description  create_app_description,
       o.create_user_id,
       cu.first_name      create_user_first_name,
       cu.last_name       create_user_last_name,
       o.create_timestamp,
       o.update_app_id,
       a2.org_id          update_app_org_id,
       a2.app_extl_id     update_app_extl_id,
       a2.app_name        update_app_name,
       a2.app_description update_app_description,
       o.update_user_id,
       uu.first_name      update_user_first_name,
       uu.last_name       update_user_last_name,
       o.update_timestamp
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
         INNER JOIN app a on a.app_id = o.create_app_id
         INNER JOIN app a2 on a2.app_id = o.update_app_id
         INNER JOIN users cu on cu.user_id = o.create_user_id
         INNER JOIN users uu on uu.user_id = o.update_user_id
WHERE o.org_extl_id = $1;

-- name: FindOrgByName :one
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       o.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
WHERE o.org_name = $1;

-- name: FindOrgByNameWithAudit :one
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       o.create_app_id,
       a.org_id           create_app_org_id,
       a.app_extl_id      create_app_extl_id,
       a.app_name         create_app_name,
       a.app_description  create_app_description,
       o.create_user_id,
       cu.first_name      create_user_first_name,
       cu.last_name       create_user_last_name,
       o.create_timestamp,
       o.update_app_id,
       a2.org_id          update_app_org_id,
       a2.app_extl_id     update_app_extl_id,
       a2.app_name        update_app_name,
       a2.app_description update_app_description,
       o.update_user_id,
       uu.first_name      update_user_first_name,
       uu.last_name       update_user_last_name,
       o.update_timestamp
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
         INNER JOIN app a on a.app_id = o.create_app_id
         INNER JOIN app a2 on a2.app_id = o.update_app_id
         INNER JOIN users cu on cu.user_id = o.create_user_id
         INNER JOIN users uu on uu.user_id = o.update_user_id
WHERE o.org_name = $1;

-- name: FindOrgs :many
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       o.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
ORDER BY org_name;

-- name: FindOrgsWithAudit :many
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_id,
       ok.org_kind_extl_id,
       ok.org_kind_desc,
       o.create_app_id,
       a.org_id           create_app_org_id,
       a.app_extl_id      create_app_extl_id,
       a.app_name         create_app_name,
       a.app_description  create_app_description,
       o.create_user_id,
       cu.first_name      create_user_first_name,
       cu.last_name       create_user_last_name,
       o.create_timestamp,
       o.update_app_id,
       a2.org_id          update_app_org_id,
       a2.app_extl_id     update_app_extl_id,
       a2.app_name        update_app_name,
       a2.app_description update_app_description,
       o.update_user_id,
       uu.first_name      update_user_first_name,
       uu.last_name       update_user_last_name,
       o.update_timestamp
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
         INNER JOIN app a on a.app_id = o.create_app_id
         INNER JOIN app a2 on a2.app_id = o.update_app_id
         INNER JOIN users cu on cu.user_id = o.create_user_id
         INNER JOIN users uu on uu.user_id = o.update_user_id;

-- name: FindOrgsByKindExtlID :many
SELECT o.org_id,
       o.org_extl_id,
       o.org_name,
       o.org_description,
       ok.org_kind_extl_id,
       ok.org_kind_desc
FROM org o
         INNER JOIN org_kind ok on ok.org_kind_id = o.org_kind_id
WHERE ok.org_kind_extl_id = $1;


-- name: CreateOrg :execrows
INSERT INTO org (org_id, org_extl_id, org_name, org_description, org_kind_id, create_app_id, create_user_id,
                 create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: UpdateOrg :execrows
UPDATE org
SET org_name         = $1,
    org_description  = $2,
    update_app_id    = $3,
    update_user_id   = $4,
    update_timestamp = $5
WHERE org_id = $6;

-- name: DeleteOrg :execrows
DELETE
FROM org
WHERE org_id = $1;

-- ---------------------------------------------------------------------------------------------------------------------
-- Org Kind
-- ---------------------------------------------------------------------------------------------------------------------

-- name: FindOrgKinds :many
SELECT *
FROM org_kind;

-- name: FindOrgKindByExtlID :one
SELECT *
FROM org_kind
WHERE org_kind_extl_id = $1;

-- name: CreateOrgKind :execrows
insert into org_kind (org_kind_id, org_kind_extl_id, org_kind_desc, create_app_id, create_user_id, create_timestamp,
                      update_app_id, update_user_id, update_timestamp)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);
