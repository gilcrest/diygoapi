-- name: CreatePermission :execrows
insert into permission (permission_id, permission_extl_id, resource, operation, permission_description, active, create_app_id,
                        create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: FindAllPermissions :many
select *
from permission;

-- name: FindPermissionByExternalID :one
SELECT *
FROM permission
WHERE permission_extl_id = $1;

-- name: FindPermissionByResourceOperation :one
SELECT *
FROM permission
WHERE resource = $1
  AND operation = $2;


-- name: CreateRole :execrows
insert into role (role_id, role_extl_id, role_cd, role_description, active, create_app_id, create_user_id,
                  create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: CreateRolePermission :execrows
insert into role_permission (role_id, permission_id, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: CreateRoleUser :execrows
insert into role_user (role_id, user_id, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: IsAuthorized :one
SELECT ru.user_id
FROM role_user ru
         INNER JOIN role_permission rp on rp.role_id = ru.role_id
         INNER JOIN permission p on p.permission_id = rp.permission_id
WHERE p.active = true
  AND p.resource = $1
  AND p.operation = $2
  AND ru.user_id = $3;
