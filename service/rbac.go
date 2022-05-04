package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/datastore/authstore"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/secure"
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

// PermissionService is a service for creating, reading, updating and deleting a Permission
type PermissionService struct {
	Datastorer Datastorer
}

// Create is used to create a Permission
func (s PermissionService) Create(ctx context.Context, r *auth.Permission, adt audit.Audit) (p auth.Permission, err error) {
	// set Unique ID for Permission
	r.ID = uuid.New()

	r.ExternalID = secure.NewID()

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

	arg := authstore.CreatePermissionParams{
		PermissionID:          r.ID,
		PermissionExtlID:      r.ExternalID.String(),
		Resource:              r.Resource,
		Operation:             r.Operation,
		PermissionDescription: r.Description,
		Active:                sql.NullBool{Bool: r.Active, Valid: true},
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

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return auth.Permission{}, err
	}

	return *r, nil
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
			Active:      row.Active.Bool,
		}
		sp = append(sp, p)
	}

	return sp, nil
}

type RoleService struct {
	Datastorer Datastorer
}

func (s RoleService) Create(ctx context.Context, r *auth.Role, adt audit.Audit) (role auth.Role, err error) {
	// set Unique ID for Role
	r.ID = uuid.New()

	// set Unique External ID
	r.ExternalID = secure.NewID()

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

	arg := authstore.CreateRoleParams{
		RoleID:          r.ID,
		RoleExtlID:      r.ExternalID.String(),
		RoleCd:          r.Code,
		Active:          sql.NullBool{Bool: r.Active, Valid: true},
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
		return auth.Role{}, err
	}

	// should only impact exactly one record
	if rowsAffected != 1 {
		return auth.Role{}, errs.E(errs.Database, fmt.Sprintf("Create() should insert 1 row, actual: %d", rowsAffected))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return auth.Role{}, err
	}

	return *r, nil
}
