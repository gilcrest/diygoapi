-- name: CreatePermission :execrows
insert into permission (permission_id, permission_extl_id, resource, operation, permission_description, active,
                        create_app_id,
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

-- name: DeletePermissionByExternalID :execrows
DELETE FROM permission
WHERE permission_extl_id = $1;

-- name: CreateRole :execrows
insert into role (role_id, role_extl_id, role_cd, role_description, active, create_app_id, create_user_id,
                  create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: FindRoleByCode :one
SELECT *
FROM role
WHERE role_cd = $1;

-- name: FindRolePermissionsByRoleID :many
SELECT p.*
FROM role_permission r
         inner join permission p on p.permission_id = r.permission_id
WHERE r.role_id = $1;


-- name: CreateRolePermission :execrows
insert into role_permission (role_id, permission_id, create_app_id, create_user_id, create_timestamp, update_app_id,
                             update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: DeleteAllPermissions4Role :execrows
DELETE FROM role_permission
WHERE role_id = $1;

-- name: CreateUsersRole :execrows
insert into users_role (user_id, role_id, org_id, create_app_id, create_user_id, create_timestamp, update_app_id,
                        update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: IsAuthorized :one
-- IsAuthorized selects a user_id which is authorized for access to a resource.
-- The query can return multiple results, but since QueryRow is used, only the first
-- is returned.
SELECT ur.user_id
FROM users_role ur
         INNER JOIN role_permission rp on rp.role_id = ur.role_id
         INNER JOIN permission p on p.permission_id = rp.permission_id
WHERE p.active = true
  AND p.resource = $1
  AND p.operation = $2
  AND ur.user_id = $3
  AND ur.org_id = $4;

-- name: FindUsersByOrgRole :many
SELECT *
FROM users_role ur
WHERE ur.org_id = $1
  aND ur.role_id = $2;


-- name: CreateAuth :execrows
INSERT INTO auth (auth_id, user_id, auth_provider_id, auth_provider_cd, auth_provider_client_id,
                  auth_provider_person_id,
                  auth_provider_access_token, auth_provider_refresh_token, auth_provider_access_token_expiry,
                  create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id,
                  update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: FindAuthByAccessToken :one
SELECT *
FROM auth
WHERE auth_provider_access_token = $1;

-- name: FindAuthByProviderUserID :one
SELECT *
FROM auth
WHERE auth_provider_id = $1
  AND auth_provider_person_id = $2;

-- name: CreateAuthProvider :execrows
INSERT INTO auth_provider (auth_provider_id, auth_provider_cd, auth_provider_desc, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);
