// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: auth.sql

package datastore

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createAuth = `-- name: CreateAuth :execrows
INSERT INTO auth (auth_id, user_id, auth_provider_id, auth_provider_cd, auth_provider_client_id,
                  auth_provider_person_id,
                  auth_provider_access_token, auth_provider_refresh_token, auth_provider_access_token_expiry,
                  create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id,
                  update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
`

type CreateAuthParams struct {
	AuthID                        pgtype.UUID
	UserID                        pgtype.UUID
	AuthProviderID                int64
	AuthProviderCd                string
	AuthProviderClientID          pgtype.Text
	AuthProviderPersonID          string
	AuthProviderAccessToken       string
	AuthProviderRefreshToken      pgtype.Text
	AuthProviderAccessTokenExpiry pgtype.Timestamptz
	CreateAppID                   pgtype.UUID
	CreateUserID                  pgtype.UUID
	CreateTimestamp               pgtype.Timestamptz
	UpdateAppID                   pgtype.UUID
	UpdateUserID                  pgtype.UUID
	UpdateTimestamp               pgtype.Timestamptz
}

func (q *Queries) CreateAuth(ctx context.Context, arg CreateAuthParams) (int64, error) {
	result, err := q.db.Exec(ctx, createAuth,
		arg.AuthID,
		arg.UserID,
		arg.AuthProviderID,
		arg.AuthProviderCd,
		arg.AuthProviderClientID,
		arg.AuthProviderPersonID,
		arg.AuthProviderAccessToken,
		arg.AuthProviderRefreshToken,
		arg.AuthProviderAccessTokenExpiry,
		arg.CreateAppID,
		arg.CreateUserID,
		arg.CreateTimestamp,
		arg.UpdateAppID,
		arg.UpdateUserID,
		arg.UpdateTimestamp,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const createAuthProvider = `-- name: CreateAuthProvider :execrows
INSERT INTO auth_provider (auth_provider_id, auth_provider_cd, auth_provider_desc, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type CreateAuthProviderParams struct {
	AuthProviderID   int64
	AuthProviderCd   string
	AuthProviderDesc string
	CreateAppID      pgtype.UUID
	CreateUserID     pgtype.UUID
	CreateTimestamp  pgtype.Timestamptz
	UpdateAppID      pgtype.UUID
	UpdateUserID     pgtype.UUID
	UpdateTimestamp  pgtype.Timestamptz
}

func (q *Queries) CreateAuthProvider(ctx context.Context, arg CreateAuthProviderParams) (int64, error) {
	result, err := q.db.Exec(ctx, createAuthProvider,
		arg.AuthProviderID,
		arg.AuthProviderCd,
		arg.AuthProviderDesc,
		arg.CreateAppID,
		arg.CreateUserID,
		arg.CreateTimestamp,
		arg.UpdateAppID,
		arg.UpdateUserID,
		arg.UpdateTimestamp,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const createPermission = `-- name: CreatePermission :execrows
insert into permission (permission_id, permission_extl_id, resource, operation, permission_description, active,
                        create_app_id,
                        create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
`

type CreatePermissionParams struct {
	PermissionID          pgtype.UUID
	PermissionExtlID      string
	Resource              string
	Operation             string
	PermissionDescription string
	Active                bool
	CreateAppID           pgtype.UUID
	CreateUserID          pgtype.UUID
	CreateTimestamp       pgtype.Timestamptz
	UpdateAppID           pgtype.UUID
	UpdateUserID          pgtype.UUID
	UpdateTimestamp       pgtype.Timestamptz
}

func (q *Queries) CreatePermission(ctx context.Context, arg CreatePermissionParams) (int64, error) {
	result, err := q.db.Exec(ctx, createPermission,
		arg.PermissionID,
		arg.PermissionExtlID,
		arg.Resource,
		arg.Operation,
		arg.PermissionDescription,
		arg.Active,
		arg.CreateAppID,
		arg.CreateUserID,
		arg.CreateTimestamp,
		arg.UpdateAppID,
		arg.UpdateUserID,
		arg.UpdateTimestamp,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const createRole = `-- name: CreateRole :execrows
insert into role (role_id, role_extl_id, role_cd, role_description, active, create_app_id, create_user_id,
                  create_timestamp, update_app_id, update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`

type CreateRoleParams struct {
	RoleID          pgtype.UUID
	RoleExtlID      string
	RoleCd          string
	RoleDescription string
	Active          bool
	CreateAppID     pgtype.UUID
	CreateUserID    pgtype.UUID
	CreateTimestamp pgtype.Timestamptz
	UpdateAppID     pgtype.UUID
	UpdateUserID    pgtype.UUID
	UpdateTimestamp pgtype.Timestamptz
}

func (q *Queries) CreateRole(ctx context.Context, arg CreateRoleParams) (int64, error) {
	result, err := q.db.Exec(ctx, createRole,
		arg.RoleID,
		arg.RoleExtlID,
		arg.RoleCd,
		arg.RoleDescription,
		arg.Active,
		arg.CreateAppID,
		arg.CreateUserID,
		arg.CreateTimestamp,
		arg.UpdateAppID,
		arg.UpdateUserID,
		arg.UpdateTimestamp,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const createRolePermission = `-- name: CreateRolePermission :execrows
insert into role_permission (role_id, permission_id, create_app_id, create_user_id, create_timestamp, update_app_id,
                             update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

type CreateRolePermissionParams struct {
	RoleID          pgtype.UUID
	PermissionID    pgtype.UUID
	CreateAppID     pgtype.UUID
	CreateUserID    pgtype.UUID
	CreateTimestamp pgtype.Timestamptz
	UpdateAppID     pgtype.UUID
	UpdateUserID    pgtype.UUID
	UpdateTimestamp pgtype.Timestamptz
}

func (q *Queries) CreateRolePermission(ctx context.Context, arg CreateRolePermissionParams) (int64, error) {
	result, err := q.db.Exec(ctx, createRolePermission,
		arg.RoleID,
		arg.PermissionID,
		arg.CreateAppID,
		arg.CreateUserID,
		arg.CreateTimestamp,
		arg.UpdateAppID,
		arg.UpdateUserID,
		arg.UpdateTimestamp,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const createUsersRole = `-- name: CreateUsersRole :execrows
insert into users_role (user_id, role_id, org_id, create_app_id, create_user_id, create_timestamp, update_app_id,
                        update_user_id, update_timestamp)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type CreateUsersRoleParams struct {
	UserID          pgtype.UUID
	RoleID          pgtype.UUID
	OrgID           pgtype.UUID
	CreateAppID     pgtype.UUID
	CreateUserID    pgtype.UUID
	CreateTimestamp pgtype.Timestamptz
	UpdateAppID     pgtype.UUID
	UpdateUserID    pgtype.UUID
	UpdateTimestamp pgtype.Timestamptz
}

func (q *Queries) CreateUsersRole(ctx context.Context, arg CreateUsersRoleParams) (int64, error) {
	result, err := q.db.Exec(ctx, createUsersRole,
		arg.UserID,
		arg.RoleID,
		arg.OrgID,
		arg.CreateAppID,
		arg.CreateUserID,
		arg.CreateTimestamp,
		arg.UpdateAppID,
		arg.UpdateUserID,
		arg.UpdateTimestamp,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const deleteAllPermissions4Role = `-- name: DeleteAllPermissions4Role :execrows
DELETE FROM role_permission
WHERE role_id = $1
`

func (q *Queries) DeleteAllPermissions4Role(ctx context.Context, roleID pgtype.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, deleteAllPermissions4Role, roleID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const deletePermissionByExternalID = `-- name: DeletePermissionByExternalID :execrows
DELETE FROM permission
WHERE permission_extl_id = $1
`

func (q *Queries) DeletePermissionByExternalID(ctx context.Context, permissionExtlID string) (int64, error) {
	result, err := q.db.Exec(ctx, deletePermissionByExternalID, permissionExtlID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const findAllPermissions = `-- name: FindAllPermissions :many
select permission_id, permission_extl_id, resource, operation, permission_description, active, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp
from permission
`

func (q *Queries) FindAllPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := q.db.Query(ctx, findAllPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Permission
	for rows.Next() {
		var i Permission
		if err := rows.Scan(
			&i.PermissionID,
			&i.PermissionExtlID,
			&i.Resource,
			&i.Operation,
			&i.PermissionDescription,
			&i.Active,
			&i.CreateAppID,
			&i.CreateUserID,
			&i.CreateTimestamp,
			&i.UpdateAppID,
			&i.UpdateUserID,
			&i.UpdateTimestamp,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findAuthByAccessToken = `-- name: FindAuthByAccessToken :one
SELECT auth_id, user_id, auth_provider_id, auth_provider_cd, auth_provider_client_id, auth_provider_person_id, auth_provider_access_token, auth_provider_refresh_token, auth_provider_access_token_expiry, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp
FROM auth
WHERE auth_provider_access_token = $1
`

func (q *Queries) FindAuthByAccessToken(ctx context.Context, authProviderAccessToken string) (Auth, error) {
	row := q.db.QueryRow(ctx, findAuthByAccessToken, authProviderAccessToken)
	var i Auth
	err := row.Scan(
		&i.AuthID,
		&i.UserID,
		&i.AuthProviderID,
		&i.AuthProviderCd,
		&i.AuthProviderClientID,
		&i.AuthProviderPersonID,
		&i.AuthProviderAccessToken,
		&i.AuthProviderRefreshToken,
		&i.AuthProviderAccessTokenExpiry,
		&i.CreateAppID,
		&i.CreateUserID,
		&i.CreateTimestamp,
		&i.UpdateAppID,
		&i.UpdateUserID,
		&i.UpdateTimestamp,
	)
	return i, err
}

const findAuthByProviderUserID = `-- name: FindAuthByProviderUserID :one
SELECT auth_id, user_id, auth_provider_id, auth_provider_cd, auth_provider_client_id, auth_provider_person_id, auth_provider_access_token, auth_provider_refresh_token, auth_provider_access_token_expiry, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp
FROM auth
WHERE auth_provider_id = $1
  AND auth_provider_person_id = $2
`

type FindAuthByProviderUserIDParams struct {
	AuthProviderID       int64
	AuthProviderPersonID string
}

func (q *Queries) FindAuthByProviderUserID(ctx context.Context, arg FindAuthByProviderUserIDParams) (Auth, error) {
	row := q.db.QueryRow(ctx, findAuthByProviderUserID, arg.AuthProviderID, arg.AuthProviderPersonID)
	var i Auth
	err := row.Scan(
		&i.AuthID,
		&i.UserID,
		&i.AuthProviderID,
		&i.AuthProviderCd,
		&i.AuthProviderClientID,
		&i.AuthProviderPersonID,
		&i.AuthProviderAccessToken,
		&i.AuthProviderRefreshToken,
		&i.AuthProviderAccessTokenExpiry,
		&i.CreateAppID,
		&i.CreateUserID,
		&i.CreateTimestamp,
		&i.UpdateAppID,
		&i.UpdateUserID,
		&i.UpdateTimestamp,
	)
	return i, err
}

const findPermissionByExternalID = `-- name: FindPermissionByExternalID :one
SELECT permission_id, permission_extl_id, resource, operation, permission_description, active, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp
FROM permission
WHERE permission_extl_id = $1
`

func (q *Queries) FindPermissionByExternalID(ctx context.Context, permissionExtlID string) (Permission, error) {
	row := q.db.QueryRow(ctx, findPermissionByExternalID, permissionExtlID)
	var i Permission
	err := row.Scan(
		&i.PermissionID,
		&i.PermissionExtlID,
		&i.Resource,
		&i.Operation,
		&i.PermissionDescription,
		&i.Active,
		&i.CreateAppID,
		&i.CreateUserID,
		&i.CreateTimestamp,
		&i.UpdateAppID,
		&i.UpdateUserID,
		&i.UpdateTimestamp,
	)
	return i, err
}

const findPermissionByResourceOperation = `-- name: FindPermissionByResourceOperation :one
SELECT permission_id, permission_extl_id, resource, operation, permission_description, active, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp
FROM permission
WHERE resource = $1
  AND operation = $2
`

type FindPermissionByResourceOperationParams struct {
	Resource  string
	Operation string
}

func (q *Queries) FindPermissionByResourceOperation(ctx context.Context, arg FindPermissionByResourceOperationParams) (Permission, error) {
	row := q.db.QueryRow(ctx, findPermissionByResourceOperation, arg.Resource, arg.Operation)
	var i Permission
	err := row.Scan(
		&i.PermissionID,
		&i.PermissionExtlID,
		&i.Resource,
		&i.Operation,
		&i.PermissionDescription,
		&i.Active,
		&i.CreateAppID,
		&i.CreateUserID,
		&i.CreateTimestamp,
		&i.UpdateAppID,
		&i.UpdateUserID,
		&i.UpdateTimestamp,
	)
	return i, err
}

const findRoleByCode = `-- name: FindRoleByCode :one
SELECT role_id, role_extl_id, role_cd, role_description, active, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp
FROM role
WHERE role_cd = $1
`

func (q *Queries) FindRoleByCode(ctx context.Context, roleCd string) (Role, error) {
	row := q.db.QueryRow(ctx, findRoleByCode, roleCd)
	var i Role
	err := row.Scan(
		&i.RoleID,
		&i.RoleExtlID,
		&i.RoleCd,
		&i.RoleDescription,
		&i.Active,
		&i.CreateAppID,
		&i.CreateUserID,
		&i.CreateTimestamp,
		&i.UpdateAppID,
		&i.UpdateUserID,
		&i.UpdateTimestamp,
	)
	return i, err
}

const findRolePermissionsByRoleID = `-- name: FindRolePermissionsByRoleID :many
SELECT p.permission_id, p.permission_extl_id, p.resource, p.operation, p.permission_description, p.active, p.create_app_id, p.create_user_id, p.create_timestamp, p.update_app_id, p.update_user_id, p.update_timestamp
FROM role_permission r
         inner join permission p on p.permission_id = r.permission_id
WHERE r.role_id = $1
`

func (q *Queries) FindRolePermissionsByRoleID(ctx context.Context, roleID pgtype.UUID) ([]Permission, error) {
	rows, err := q.db.Query(ctx, findRolePermissionsByRoleID, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Permission
	for rows.Next() {
		var i Permission
		if err := rows.Scan(
			&i.PermissionID,
			&i.PermissionExtlID,
			&i.Resource,
			&i.Operation,
			&i.PermissionDescription,
			&i.Active,
			&i.CreateAppID,
			&i.CreateUserID,
			&i.CreateTimestamp,
			&i.UpdateAppID,
			&i.UpdateUserID,
			&i.UpdateTimestamp,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const findUsersByOrgRole = `-- name: FindUsersByOrgRole :many
SELECT user_id, role_id, org_id, create_app_id, create_user_id, create_timestamp, update_app_id, update_user_id, update_timestamp
FROM users_role ur
WHERE ur.org_id = $1
  aND ur.role_id = $2
`

type FindUsersByOrgRoleParams struct {
	OrgID  pgtype.UUID
	RoleID pgtype.UUID
}

func (q *Queries) FindUsersByOrgRole(ctx context.Context, arg FindUsersByOrgRoleParams) ([]UsersRole, error) {
	rows, err := q.db.Query(ctx, findUsersByOrgRole, arg.OrgID, arg.RoleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UsersRole
	for rows.Next() {
		var i UsersRole
		if err := rows.Scan(
			&i.UserID,
			&i.RoleID,
			&i.OrgID,
			&i.CreateAppID,
			&i.CreateUserID,
			&i.CreateTimestamp,
			&i.UpdateAppID,
			&i.UpdateUserID,
			&i.UpdateTimestamp,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const isAuthorized = `-- name: IsAuthorized :one
SELECT ur.user_id
FROM users_role ur
         INNER JOIN role_permission rp on rp.role_id = ur.role_id
         INNER JOIN permission p on p.permission_id = rp.permission_id
WHERE p.active = true
  AND p.resource = $1
  AND p.operation = $2
  AND ur.user_id = $3
  AND ur.org_id = $4
`

type IsAuthorizedParams struct {
	Resource  string
	Operation string
	UserID    pgtype.UUID
	OrgID     pgtype.UUID
}

// IsAuthorized selects a user_id which is authorized for access to a resource.
// The query can return multiple results, but since QueryRow is used, only the first
// is returned.
func (q *Queries) IsAuthorized(ctx context.Context, arg IsAuthorizedParams) (pgtype.UUID, error) {
	row := q.db.QueryRow(ctx, isAuthorized,
		arg.Resource,
		arg.Operation,
		arg.UserID,
		arg.OrgID,
	)
	var user_id pgtype.UUID
	err := row.Scan(&user_id)
	return user_id, err
}
