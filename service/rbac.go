package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diy-go-api/datastore/authstore"
	"github.com/gilcrest/diy-go-api/datastore/userstore"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/auth"
	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/gilcrest/diy-go-api/domain/user"
)

// DBAuthorizer determines authorization for a user
// by running sql against tables in the database
type DBAuthorizer struct {
	Datastorer Datastorer
}

// Authorize ensures that a subject (user.User) can perform a
// particular action on a resource, e.g. subject otto.maddox711@gmail.com
// can read (GET) the resource /api/v1/movies (path).
//
// The http.Request context is used to determine the route/path information
// and must be issued through the gorilla/mux library.
func (a DBAuthorizer) Authorize(lgr zerolog.Logger, r *http.Request, adt audit.Audit) error {

	// current matched route for the request
	route := mux.CurrentRoute(r)

	// CurrentRoute can return a nil if route not setup properly or
	// is being called outside the handler of the matched route
	if route == nil {
		return errs.E(errs.Unauthorized, "nil route returned from mux.CurrentRoute")
	}

	pathTemplate, err := route.GetPathTemplate()
	if err != nil {
		return errs.E(errs.Unauthorized, err)
	}

	arg := authstore.IsAuthorizedParams{
		Resource:  pathTemplate,
		Operation: r.Method,
		UserID:    adt.User.ID,
	}

	// call IsAuthorized method to validate user has access to the resource and operation
	var authorizedID uuid.UUID
	authorizedID, err = authstore.New(a.Datastorer.Pool()).IsAuthorized(r.Context(), arg)
	if err != nil || authorizedID == uuid.Nil {
		lgr.Info().Str("user", adt.User.Username).Str("resource", pathTemplate).Str("operation", r.Method).
			Msgf("Unauthorized (user: %s, resource: %s, operation: %s)", adt.User.Username, pathTemplate, r.Method)

		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// If the user has gotten here, they have gotten through authentication
		// but do have the right access, this they are Unauthorized
		return errs.E(errs.Unauthorized, fmt.Sprintf("user %s does not have %s permission for %s", adt.User.Username, r.Method, pathTemplate))
	}

	lgr.Debug().Str("user", adt.User.Username).Str("resource", pathTemplate).Str("operation", r.Method).
		Msgf("Authorized (user: %s, resource: %s, operation: %s)", adt.User.Username, pathTemplate, r.Method)

	return nil
}

// PermissionRequest is the request struct for creating a permission
type PermissionRequest struct {
	// Unique External ID to be given to outside callers.
	ExternalID string `json:"external_id"`
	// A human-readable string which represents a resource (e.g. an HTTP route or document, etc.).
	Resource string `json:"resource"`
	// A string representing the action taken on the resource (e.g. POST, GET, edit, etc.)
	Operation string `json:"operation"`
	// A description of what the permission is granting, e.g. "grants ability to edit a billing document".
	Description string `json:"description"`
	// A boolean denoting whether the permission is active (true) or not (false).
	Active bool `json:"active"`
}

// PermissionService is a service for creating, reading, updating and deleting a Permission
type PermissionService struct {
	Datastorer Datastorer
}

// Create is used to create a Permission
func (s PermissionService) Create(ctx context.Context, r *PermissionRequest, adt audit.Audit) (p auth.Permission, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return auth.Permission{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	p, err = createPermissionTx(ctx, tx, r, adt)
	if err != nil {
		return auth.Permission{}, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return auth.Permission{}, err
	}

	return p, nil
}

// createPermissionTX separates the transaction logic as it needs to also be called during Genesis
func createPermissionTx(ctx context.Context, tx pgx.Tx, r *PermissionRequest, adt audit.Audit) (p auth.Permission, err error) {
	p = auth.Permission{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Resource:    r.Resource,
		Operation:   r.Operation,
		Description: r.Description,
		Active:      r.Active,
	}

	err = p.IsValid()
	if err != nil {
		return auth.Permission{}, err
	}

	arg := authstore.CreatePermissionParams{
		PermissionID:          p.ID,
		PermissionExtlID:      p.ExternalID.String(),
		Resource:              p.Resource,
		Operation:             p.Operation,
		PermissionDescription: p.Description,
		Active:                p.Active,
		CreateAppID:           adt.App.ID,
		CreateUserID:          adt.User.NullUUID(),
		CreateTimestamp:       time.Now(),
		UpdateAppID:           adt.App.ID,
		UpdateUserID:          adt.User.NullUUID(),
		UpdateTimestamp:       time.Now(),
	}

	var rowsAffected int64
	rowsAffected, err = authstore.New(tx).CreatePermission(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return auth.Permission{}, errs.E(errs.Exist, errs.Exist.String())
			}
			return auth.Permission{}, errs.E(errs.Database, pgErr.Message)
		}
		return auth.Permission{}, errs.E(errs.Database, err)
	}

	// should only impact exactly one record
	if rowsAffected != 1 {
		return auth.Permission{}, errs.E(errs.Database, fmt.Sprintf("Create() should insert 1 row, actual: %d", rowsAffected))
	}

	return p, nil
}

// FindAll retrieves all permissions
func (s PermissionService) FindAll(ctx context.Context) ([]auth.Permission, error) {

	rows, err := authstore.New(s.Datastorer.Pool()).FindAllPermissions(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	var sp []auth.Permission
	for _, row := range rows {
		p := auth.Permission{
			ID:          row.PermissionID,
			ExternalID:  secure.MustParseIdentifier(row.PermissionExtlID),
			Resource:    row.Resource,
			Operation:   row.Operation,
			Description: row.PermissionDescription,
			Active:      row.Active,
		}
		sp = append(sp, p)
	}

	return sp, nil
}

// newPermission initializes an auth.Permission given an authstore.Permission
func newPermission(ap authstore.Permission) auth.Permission {
	return auth.Permission{
		ID:          ap.PermissionID,
		ExternalID:  secure.MustParseIdentifier(ap.PermissionExtlID),
		Resource:    ap.Resource,
		Operation:   ap.Operation,
		Description: ap.PermissionDescription,
		Active:      ap.Active,
	}
}

// CreateRoleRequest is the request struct for creating a role
type CreateRoleRequest struct {
	// A human-readable code which represents the role.
	Code string `json:"role_cd"`
	// A longer description of the role.
	Description string `json:"role_description"`
	// A boolean denoting whether the role is active (true) or not (false).
	Active bool `json:"active"`
	// The list of permissions to be given to the role
	Permissions   []PermissionRequest
	UserExternals []string
}

// RoleService is a service for creating, reading, updating and deleting a Role
type RoleService struct {
	Datastorer Datastorer
}

// Create is used to create a Role
func (s RoleService) Create(ctx context.Context, r *CreateRoleRequest, adt audit.Audit) (role auth.Role, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return auth.Role{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	role, err = createRoleTx(ctx, tx, r, adt)
	if err != nil {
		return auth.Role{}, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return auth.Role{}, err
	}

	return role, nil
}

// createRoleTx separates out the transaction logic for creating a role as it needs to be called in multiple places
func createRoleTx(ctx context.Context, tx pgx.Tx, r *CreateRoleRequest, adt audit.Audit) (role auth.Role, err error) {

	var rolePermissions []auth.Permission
	rolePermissions, err = findPermissionsForRole(ctx, tx, r.Permissions)
	if err != nil {
		return auth.Role{}, err
	}

	var roleUsers []user.User
	roleUsers, err = findUsersForRole(ctx, tx, r.UserExternals)
	if err != nil {
		return auth.Role{}, err
	}

	role = auth.Role{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Code:        r.Code,
		Description: r.Description,
		Active:      r.Active,
		Permissions: rolePermissions,
		Users:       roleUsers,
	}

	err = role.IsValid()
	if err != nil {
		return auth.Role{}, err
	}

	arg := authstore.CreateRoleParams{
		RoleID:          role.ID,
		RoleExtlID:      role.ExternalID.String(),
		RoleCd:          role.Code,
		Active:          role.Active,
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: time.Now(),
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: time.Now(),
	}

	var rowsAffected int64
	rowsAffected, err = authstore.New(tx).CreateRole(ctx, arg)
	if err != nil {
		return auth.Role{}, errs.E(errs.Database, err)
	}

	// should only impact exactly one record
	if rowsAffected != 1 {
		return auth.Role{}, errs.E(errs.Database, fmt.Sprintf("Create() should insert 1 row, actual: %d", rowsAffected))
	}

	for _, rp := range role.Permissions {
		t := time.Now()
		createRolePermissionParams := authstore.CreateRolePermissionParams{
			RoleID:          role.ID,
			PermissionID:    rp.ID,
			CreateAppID:     adt.App.ID,
			CreateUserID:    adt.User.NullUUID(),
			CreateTimestamp: t,
			UpdateAppID:     adt.App.ID,
			UpdateUserID:    adt.User.NullUUID(),
			UpdateTimestamp: t,
		}

		rowsAffected, err = authstore.New(tx).CreateRolePermission(ctx, createRolePermissionParams)
		if err != nil {
			return auth.Role{}, errs.E(errs.Database, err)
		}

		// should only impact exactly one record
		if rowsAffected != 1 {
			return auth.Role{}, errs.E(errs.Database, fmt.Sprintf("Create() should insert 1 row, actual: %d", rowsAffected))
		}
	}

	for _, ru := range role.Users {
		t := time.Now()
		createRoleUserParams := authstore.CreateRoleUserParams{
			RoleID:          role.ID,
			UserID:          ru.ID,
			CreateAppID:     adt.App.ID,
			CreateUserID:    adt.User.NullUUID(),
			CreateTimestamp: t,
			UpdateAppID:     adt.App.ID,
			UpdateUserID:    adt.User.NullUUID(),
			UpdateTimestamp: t,
		}
		rowsAffected, err = authstore.New(tx).CreateRoleUser(ctx, createRoleUserParams)
		if err != nil {
			return auth.Role{}, errs.E(errs.Database, err)
		}

		// should only impact exactly one record
		if rowsAffected != 1 {
			return auth.Role{}, errs.E(errs.Database, fmt.Sprintf("Create() should insert 1 row, actual: %d", rowsAffected))
		}
	}

	return role, nil
}

func findPermissionsForRole(ctx context.Context, tx pgx.Tx, prs []PermissionRequest) (aps []auth.Permission, err error) {

	// it's fine for zero permissions to be added as part of a role
	if len(prs) == 0 {
		return nil, nil
	}

	// if permissions are set as part of role create, find them in the db depending on
	// which key is sent (external id or resource/operation)
	for _, pr := range prs {
		var ap authstore.Permission
		if pr.ExternalID != "" {
			ap, err = authstore.New(tx).FindPermissionByExternalID(ctx, pr.ExternalID)
			if err != nil {
				return nil, errs.E(errs.Database, err)
			}
			aps = append(aps, newPermission(ap))
		} else {
			ap, err = authstore.New(tx).FindPermissionByResourceOperation(ctx, authstore.FindPermissionByResourceOperationParams{Resource: pr.Resource, Operation: pr.Operation})
			if err != nil {
				return nil, errs.E(errs.Database, err)
			}
			aps = append(aps, newPermission(ap))
		}
	}

	return aps, nil
}

func findUsersForRole(ctx context.Context, tx pgx.Tx, extls []string) (users []user.User, err error) {

	// it's fine for zero users to be added when creating a role
	if len(extls) == 0 {
		return nil, nil
	}

	// if users are set as part of role create, find them in the db depending on
	// which key is sent (external id or resource/operation)
	for _, s := range extls {
		var row userstore.FindUserByExternalIDRow
		row, err = userstore.New(tx).FindUserByExternalID(ctx, s)
		if err != nil {
			return nil, errs.E(errs.Database, err)
		}
		users = append(users, hydrateUserFromExternalIDRow(row))
	}

	return users, nil
}
