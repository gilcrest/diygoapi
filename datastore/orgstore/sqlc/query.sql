-- name: FindOrgKinds :many
SELECT * FROM org_kind;

-- name: FindOrgKindByExtlID :one
SELECT * FROM org_kind
WHERE org_kind_extl_id = $1;

-- name: CreateOrgKind :execresult
insert into org_kind (org_kind_id, org_kind_extl_id, org_kind_desc, create_app_id, create_user_id, create_timestamp,
                      update_app_id, update_user_id, update_timestamp)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: FindOrgByKindExtlID :many
SELECT o.* FROM org o
                    INNER JOIN org_kind ot on ot.org_kind_id = o.org_kind_id
WHERE ot.org_kind_extl_id = $1;

-- name: FindOrgByID :one
SELECT * FROM org
WHERE org_id = $1 LIMIT 1;

-- name: FindOrgByExtlID :one
SELECT * FROM org
WHERE org_extl_id = $1 LIMIT 1;

-- name: FindOrgs :many
SELECT * FROM org
ORDER BY org_name;

-- name: CreateOrg :execresult
INSERT INTO org (org_id, org_extl_id, org_name, org_description, org_kind_id, create_app_id, create_user_id,
                 create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: UpdateOrg :exec
UPDATE org
SET org_name         = $1,
    org_description  = $2,
    update_app_id    = $3,
    update_user_id   = $4,
    update_timestamp = $5
WHERE org_id = $6;

-- name: DeleteOrg :exec
DELETE FROM org
WHERE org_id = $1;
