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
	"golang.org/x/oauth2"
	"golang.org/x/text/language"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
	"github.com/gilcrest/diygoapi/sqldb/datastore"
)

// DBAuthenticationService is a service which manages Oauth2 authentication
// using the database.
type DBAuthenticationService struct {
	Datastorer      diygoapi.Datastorer
	TokenExchanger  diygoapi.TokenExchanger
	EncryptionKey   *[32]byte
	LanguageMatcher language.Matcher
}

// FindAuth searches for an existing Auth object in the datastore.
//
// If an auth object already exists in the datastore for the oauth2.AccessToken
// and the oauth2.AccessToken is not past its expiration date, that auth is returned.
//
// If no auth object exists in the datastore for the access token, an attempt
// will be made to find the user's auth with the provider id and unique ID
// given by the provider (found by calling the provider API). If an auth
// object exists, it will be updated with the new access token details.
//
// The returned app and user as part of the auth object from either scenario
// above will be set to the request context for downstream use. The only
// exception is if an app is already set to the request context from upstream
// authentication, in which case, the upstream app overrides the app derived
// from the Oauth2 provider.
func (s DBAuthenticationService) FindAuth(ctx context.Context, params diygoapi.AuthenticationParams) (auth diygoapi.Auth, err error) {
	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return diygoapi.Auth{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	auth, err = findAuthByAccessToken(ctx, tx, params)
	if err != nil {
		// if error is something other than NotExist, then return error
		if !errs.KindIs(errs.NotExist, err) {
			return diygoapi.Auth{}, err
		}

		// auth could not be found by access token in the db
		// get ProviderInfo from provider API
		var providerInfo *diygoapi.ProviderInfo
		providerInfo, err = s.TokenExchanger.Exchange(ctx, params.Realm, params.Provider, params.Token)
		if err != nil {
			return diygoapi.Auth{}, err
		}

		fParams := findAuthByProviderExternalIDParams{
			Realm:        params.Realm,
			ProviderInfo: providerInfo,
			Token:        params.Token,
		}

		// search by Provider External ID
		auth, err = findAuthByProviderExternalID(ctx, tx, fParams)
		if err != nil {
			if errs.KindIs(errs.NotExist, err) {
				return diygoapi.Auth{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), err)
			}
			return diygoapi.Auth{}, err
		}
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return diygoapi.Auth{}, err
	}

	return auth, nil
}

// findAuthByAccessToken looks up an authentication object (Auth)
// given an Access Token. If found, check if there is an app
// present in the request context. If an app exists and matches the
// app stored in the auth object from the datastore, use Auth as is.
// If they are different, update the auth object in the datastore
// with the app in the context. If an app is set to the context already,
// the app is an internally created app and overrides the app given by
// the authentication provider.
//
// If none are found, an error with errs.NotExist kind is returned.
func findAuthByAccessToken(ctx context.Context, tx pgx.Tx, params diygoapi.AuthenticationParams) (diygoapi.Auth, error) {

	var (
		dbAuth datastore.Auth
		err    error
	)

	// determine if there is already an auth record created in the db
	// using the given access token.
	//
	// If no record exists in the database, or a database error occurs,
	// return the appropriate error.
	dbAuth, err = datastore.New(tx).FindAuthByAccessToken(ctx, params.Token.AccessToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return diygoapi.Auth{}, errs.E(errs.NotExist, "No auth found in db for access token")
		} else {
			return diygoapi.Auth{}, errs.E(errs.Database, err)
		}
	}

	// populate Person
	var u *diygoapi.User
	u, err = FindUserByID(ctx, tx, dbAuth.UserID)
	if err != nil {
		return diygoapi.Auth{}, err
	}

	// populate Auth
	auth := diygoapi.Auth{
		ID:               dbAuth.AuthID,
		User:             u,
		Provider:         diygoapi.Provider(dbAuth.AuthProviderID),
		ProviderClientID: dbAuth.AuthProviderClientID.String,
		ProviderPersonID: dbAuth.AuthProviderPersonID,
		Token: &oauth2.Token{
			AccessToken:  dbAuth.AuthProviderAccessToken,
			TokenType:    diygoapi.BearerTokenType,
			RefreshToken: dbAuth.AuthProviderRefreshToken.String,
			Expiry:       dbAuth.AuthProviderAccessTokenExpiry.Time},
	}

	// if token is no longer valid, return an error
	if !auth.Token.Valid() {
		return diygoapi.Auth{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "token is no longer valid")
	}

	return auth, nil
}

type findAuthByProviderExternalIDParams struct {
	Realm        string
	ProviderInfo *diygoapi.ProviderInfo
	Token        *oauth2.Token
}

// findAuthByProviderExternalID searches for an auth for the User using
// the authentication provider's external ID. If an auth object exists, it
// will be updated with the new access token details.
func findAuthByProviderExternalID(ctx context.Context, tx pgx.Tx, params findAuthByProviderExternalIDParams) (diygoapi.Auth, error) {
	var err error

	findAuthByProviderUserIDParams := datastore.FindAuthByProviderUserIDParams{
		AuthProviderID:       int64(params.ProviderInfo.Provider),
		AuthProviderPersonID: params.ProviderInfo.UserInfo.ExternalID,
	}

	// find the user's auth record by Provider and Provider Unique ID
	var dbAuth datastore.Auth
	dbAuth, err = datastore.New(tx).FindAuthByProviderUserID(ctx, findAuthByProviderUserIDParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return diygoapi.Auth{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), fmt.Sprintf("no authorization object, Provider: %s, Provider Person ID: %s, email: %s", params.ProviderInfo.Provider.String(), params.ProviderInfo.UserInfo.ExternalID, params.ProviderInfo.UserInfo.Email))
		} else {
			return diygoapi.Auth{}, errs.E(errs.Database, err)
		}
	}

	// populate User
	var u *diygoapi.User
	u, err = FindUserByID(ctx, tx, dbAuth.UserID)
	if err != nil {
		return diygoapi.Auth{}, err
	}

	// populate Auth
	auth := diygoapi.Auth{
		ID:               dbAuth.AuthID,
		User:             u,
		Provider:         params.ProviderInfo.Provider,
		ProviderClientID: params.ProviderInfo.TokenInfo.ClientID,
		ProviderPersonID: params.ProviderInfo.UserInfo.ExternalID,
		Token:            params.Token,
	}

	// if token is no longer valid, return an error
	if !auth.Token.Valid() {
		return diygoapi.Auth{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "token is no longer valid")
	}

	return auth, nil
}

// FindAppByProviderClientID finds an app given a Provider's Unique Client ID
func (s DBAuthenticationService) FindAppByProviderClientID(ctx context.Context, realm string, auth diygoapi.Auth) (a *diygoapi.App, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	a, err = findAppByProviderClientID(ctx, tx, auth.ProviderClientID)
	if err != nil {
		if errs.KindIs(errs.NotExist, err) {
			return nil, errs.E(errs.Unauthenticated, errs.Realm(realm), fmt.Sprintf("No app mapped to Client ID: %s for Provider: %s", auth.ProviderClientID, auth.Provider.String()))
		}
		return nil, errs.E(errs.Unauthenticated, errs.Realm(realm), err)
	}

	return a, nil
}

// FindAppByAPIKey finds an app given its External ID and determines
// if the given API key is a valid key for it. It is used as part of
// app authentication
func (s DBAuthenticationService) FindAppByAPIKey(ctx context.Context, realm, appExtlID, key string) (a *diygoapi.App, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var kr []datastore.FindAppAPIKeysByAppExtlIDRow

	// retrieve the list of encrypted API keys from the database
	kr, err = datastore.New(tx).FindAppAPIKeysByAppExtlID(ctx, appExtlID)
	if err != nil {
		return nil, errs.E(errs.Unauthenticated, errs.Realm(realm), err)
	}

	var (
		ak  diygoapi.APIKey
		aks []diygoapi.APIKey
	)

	a = new(diygoapi.App)

	// for each row, decrypt the API key using the encryption key,
	// initialize an app.APIKey and set to a slice of API keys.
	for i, row := range kr {
		if i == 0 { // only need to fill the app struct on first iteration
			var extl secure.Identifier
			extl, err = secure.ParseIdentifier(row.OrgExtlID)
			if err != nil {
				return nil, err
			}
			a.ID = row.AppID
			a.ExternalID = extl
			a.Org = &diygoapi.Org{
				ID:          row.OrgID,
				ExternalID:  extl,
				Name:        row.OrgName,
				Description: row.OrgDescription,
			}
			a.Name = row.AppName
			a.Description = row.AppDescription
		}
		ak, err = diygoapi.NewAPIKeyFromCipher(row.ApiKey, s.EncryptionKey)
		if err != nil {
			return nil, err
		}
		ak.SetDeactivationDate(row.DeactvDate)
		aks = append(aks, ak)
	}
	a.APIKeys = aks

	// ValidKey determines if any of the keys attached to the app
	// match the input key and are still valid.
	err = a.ValidateKey(realm, key)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// SelfRegister is used for first-time registration of a Person/User
// in the system (associated with an Organization). This is "self
// registration" as opposed to one person registering another person.
//
// SelfRegister creates an Auth object and a Person/User and stores
// them in the database. A search is done prior to creation to
// determine if user is already registered, and if so, the existing
// user is returned.
func (s DBAuthenticationService) SelfRegister(ctx context.Context, params diygoapi.AuthenticationParams) (auth diygoapi.Auth, err error) {
	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return diygoapi.Auth{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var providerInfo *diygoapi.ProviderInfo
	auth, err = findAuthByAccessToken(ctx, tx, params)
	if err != nil {
		// if error is something other than NotExist, then return error
		if !errs.KindIs(errs.NotExist, err) {
			return diygoapi.Auth{}, err
		}

		// auth could not be found by access token in the db
		// get ProviderInfo from provider API
		providerInfo, err = s.TokenExchanger.Exchange(ctx, params.Realm, params.Provider, params.Token)
		if err != nil {
			return diygoapi.Auth{}, err
		}

		fParams := findAuthByProviderExternalIDParams{
			Realm:        params.Realm,
			ProviderInfo: providerInfo,
			Token:        params.Token,
		}

		// we've gotten here, error kind is NotExist, so auth could not be found by
		// access token. Try to find auth by Provider External ID
		auth, err = findAuthByProviderExternalID(ctx, tx, fParams)
		if err != nil {
			// if error is something other than NotExist, then return error
			if !errs.KindIs(errs.NotExist, err) {
				return diygoapi.Auth{}, err
			}
		}
	}

	// if auth still has not been found (we know this by checking if auth ID is nil)
	// then create a new Auth for the User
	if auth.ID == uuid.Nil {
		var a *diygoapi.App
		// check app from context first
		a, _ = diygoapi.AppFromContext(ctx)

		// if no app in context, get app from Provider
		if a == nil {
			a, err = findAppByProviderClientID(ctx, tx, providerInfo.TokenInfo.ClientID)
			if err != nil {
				if errs.KindIs(errs.NotExist, err) {
					return diygoapi.Auth{}, errs.E(errs.NotExist, fmt.Sprintf("no app registered for Provider: %s, Client ID: %s", params.Provider.String(), providerInfo.TokenInfo.ClientID))
				}
				return diygoapi.Auth{}, err
			}
		}

		u := newUserFromProviderInfo(providerInfo, s.LanguageMatcher)

		err = u.Validate()
		if err != nil {
			return diygoapi.Auth{}, err
		}

		p := diygoapi.Person{
			ID:         uuid.New(),
			ExternalID: secure.NewID(),
			Users:      []*diygoapi.User{u},
		}

		adt := diygoapi.Audit{
			App:    a,
			User:   u,
			Moment: time.Now(),
		}

		// write Person/User from request to the database
		err = createPersonTx(ctx, tx, p, adt)
		if err != nil {
			return diygoapi.Auth{}, err
		}

		// associate user to the app's org
		aoaParams := attachOrgAssociationParams{
			Org:   a.Org,
			User:  u,
			Audit: adt,
		}
		err = attachOrgAssociation(ctx, tx, aoaParams)
		if err != nil {
			return diygoapi.Auth{}, err
		}

		auth = diygoapi.Auth{
			ID:               uuid.New(),
			User:             u,
			Provider:         providerInfo.Provider,
			ProviderClientID: providerInfo.TokenInfo.ClientID,
			ProviderPersonID: providerInfo.UserInfo.ExternalID,
			Token:            params.Token,
		}

		err = createAuthTx(ctx, tx, createAuthTxParams{Auth: auth, Audit: adt})
		if err != nil {
			return diygoapi.Auth{}, err
		}

	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return diygoapi.Auth{}, err
	}

	return auth, nil
}

// newUserFromProviderInfo creates a new User struct to be used in db user creation
func newUserFromProviderInfo(pi *diygoapi.ProviderInfo, lm language.Matcher) *diygoapi.User {
	var langPrefs []language.Tag
	langPref, _, _ := lm.Match(language.Make(pi.UserInfo.Locale))
	langPrefs = append(langPrefs, langPref)

	// create User from ProviderInfo
	u := &diygoapi.User{
		ID:                  uuid.New(),
		ExternalID:          secure.NewID(),
		NamePrefix:          pi.UserInfo.NamePrefix,
		FirstName:           pi.UserInfo.FirstName,
		MiddleName:          pi.UserInfo.MiddleName,
		LastName:            pi.UserInfo.LastName,
		FullName:            pi.UserInfo.FullName,
		NameSuffix:          pi.UserInfo.NameSuffix,
		Nickname:            pi.UserInfo.Nickname,
		Gender:              pi.UserInfo.Gender,
		Email:               pi.UserInfo.Email,
		BirthDate:           pi.UserInfo.BirthDate,
		LanguagePreferences: langPrefs,
		HostedDomain:        pi.UserInfo.HostedDomain,
		PictureURL:          pi.UserInfo.Picture,
		ProfileLink:         pi.UserInfo.ProfileLink,
		Source:              pi.Provider.String(),
	}

	return u
}

type createAuthTxParams struct {
	Auth  diygoapi.Auth
	Audit diygoapi.Audit
}

func createAuthTx(ctx context.Context, tx pgx.Tx, params createAuthTxParams) (err error) {
	createAuthParams := datastore.CreateAuthParams{
		AuthID:                        params.Auth.ID,
		UserID:                        params.Auth.User.ID,
		AuthProviderID:                int64(params.Auth.Provider),
		AuthProviderCd:                params.Auth.Provider.String(),
		AuthProviderClientID:          diygoapi.NewNullString(params.Auth.ProviderClientID),
		AuthProviderPersonID:          params.Auth.ProviderPersonID,
		AuthProviderAccessToken:       params.Auth.Token.AccessToken,
		AuthProviderRefreshToken:      diygoapi.NewNullString(params.Auth.Token.RefreshToken),
		AuthProviderAccessTokenExpiry: diygoapi.NewNullTime(params.Auth.Token.Expiry),
		CreateAppID:                   params.Audit.App.ID,
		CreateUserID:                  params.Audit.User.NullUUID(),
		CreateTimestamp:               params.Audit.Moment,
		UpdateAppID:                   params.Audit.App.ID,
		UpdateUserID:                  params.Audit.User.NullUUID(),
		UpdateTimestamp:               params.Audit.Moment,
	}

	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).CreateAuth(ctx, createAuthParams)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	// should only create exactly one record
	if rowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("CreateAuth() should insert 1 row, actual: %d", rowsAffected))
	}

	return nil
}

// DBAuthorizationService manages authorization using the database.
type DBAuthorizationService struct {
	Datastorer diygoapi.Datastorer
}

// Authorize ensures that a subject (User) can perform a
// particular action on a resource, e.g. subject otto.maddox711@gmail.com
// can read (GET) the resource /api/v1/movies (path).
//
// The http.Request context is used to determine the route/path information
// and must be issued through the gorilla/mux library.
//
// Authorize implements Role Based Access Control (RBAC), in this case,
// determining authorization for a user by running sql against tables
// in the database
func (s *DBAuthorizationService) Authorize(r *http.Request, lgr zerolog.Logger, adt diygoapi.Audit) (err error) {
	ctx := r.Context()

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// current matched route for the request
	route := mux.CurrentRoute(r)

	// CurrentRoute can return a nil if route not setup properly or
	// is being called outside the handler of the matched route
	if route == nil {
		return errs.E(errs.Unauthorized, "nil route returned from mux.CurrentRoute")
	}

	var pathTemplate string
	pathTemplate, err = route.GetPathTemplate()
	if err != nil {
		return errs.E(errs.Unauthorized, err)
	}

	arg := datastore.IsAuthorizedParams{
		Resource:  pathTemplate,
		Operation: r.Method,
		UserID:    adt.User.ID,
		// Set the Org using the org the audit app is associated to.
		// The business assumption currently is that an app can
		// only belong to one org.
		OrgID: adt.App.Org.ID,
	}

	// call IsAuthorized method to validate user has access to the resource and operation
	var authorizedID uuid.UUID
	authorizedID, err = datastore.New(tx).IsAuthorized(r.Context(), arg)
	if err != nil || authorizedID == uuid.Nil {
		lgr.Info().Str("user_extl_id", adt.User.ExternalID.String()).Str("resource", pathTemplate).Str("operation", r.Method).
			Msgf("Unauthorized (user_extl_id: %s, resource: %s, operation: %s)", adt.User.ExternalID.String(), pathTemplate, r.Method)

		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// If the user has gotten here, they have gotten through authentication
		// but do have the right access, this they are Unauthorized
		return errs.E(errs.Unauthorized, fmt.Sprintf("User_extl_id %s does not have %s permission for %s", adt.User.ExternalID.String(), r.Method, pathTemplate))
	}

	lgr.Debug().Str("user_extl_id", adt.User.ExternalID.String()).Str("resource", pathTemplate).Str("operation", r.Method).
		Msgf("Authorized (user_extl_id: %s, resource: %s, operation: %s)", adt.User.ExternalID.String(), pathTemplate, r.Method)

	return nil
}

// PermissionService is a service for creating, reading, updating and deleting a Permission
type PermissionService struct {
	Datastorer diygoapi.Datastorer
}

// Create is used to create a Permission
func (s *PermissionService) Create(ctx context.Context, r *diygoapi.CreatePermissionRequest, adt diygoapi.Audit) (response *diygoapi.PermissionResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var p diygoapi.Permission
	p, err = createPermissionTx(ctx, tx, r, adt)
	if err != nil {
		return nil, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	response = &diygoapi.PermissionResponse{
		ExternalID:  p.ExternalID.String(),
		Resource:    p.Resource,
		Operation:   p.Operation,
		Description: p.Description,
		Active:      p.Active,
	}

	return response, nil
}

// createPermissionTX separates the transaction logic as it needs to also be called during Genesis
func createPermissionTx(ctx context.Context, tx pgx.Tx, r *diygoapi.CreatePermissionRequest, adt diygoapi.Audit) (p diygoapi.Permission, err error) {
	p = diygoapi.Permission{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Resource:    r.Resource,
		Operation:   r.Operation,
		Description: r.Description,
		Active:      r.Active,
	}

	err = p.Validate()
	if err != nil {
		return diygoapi.Permission{}, err
	}

	arg := datastore.CreatePermissionParams{
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
	rowsAffected, err = datastore.New(tx).CreatePermission(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return diygoapi.Permission{}, errs.E(errs.Exist, errs.Exist.String())
			}
			return diygoapi.Permission{}, errs.E(errs.Database, pgErr.Message)
		}
		return diygoapi.Permission{}, errs.E(errs.Database, err)
	}

	// should only impact exactly one record
	if rowsAffected != 1 {
		return diygoapi.Permission{}, errs.E(errs.Database, fmt.Sprintf("Create() should insert 1 row, actual: %d", rowsAffected))
	}

	return p, nil
}

// FindAll retrieves all permissions
func (s *PermissionService) FindAll(ctx context.Context) (permissions []*diygoapi.PermissionResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var rows []datastore.Permission
	rows, err = datastore.New(tx).FindAllPermissions(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	var sp []*diygoapi.PermissionResponse
	for _, row := range rows {
		p := &diygoapi.PermissionResponse{
			ExternalID:  row.PermissionExtlID,
			Resource:    row.Resource,
			Operation:   row.Operation,
			Description: row.PermissionDescription,
			Active:      row.Active,
		}
		sp = append(sp, p)
	}

	return sp, nil
}

// Delete is used to delete a Permission
func (s *PermissionService) Delete(ctx context.Context, extlID string) (dr diygoapi.DeleteResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return diygoapi.DeleteResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).DeletePermissionByExternalID(ctx, extlID)
	if err != nil {
		return diygoapi.DeleteResponse{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return diygoapi.DeleteResponse{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return diygoapi.DeleteResponse{}, err
	}

	response := diygoapi.DeleteResponse{
		ExternalID: extlID,
		Deleted:    true,
	}

	return response, nil
}

// newPermission initializes a Permission given a datastore.Permission
func newPermission(ap datastore.Permission) *diygoapi.Permission {
	return &diygoapi.Permission{
		ID:          ap.PermissionID,
		ExternalID:  secure.MustParseIdentifier(ap.PermissionExtlID),
		Resource:    ap.Resource,
		Operation:   ap.Operation,
		Description: ap.PermissionDescription,
		Active:      ap.Active,
	}
}

// RoleService is a service for creating, reading, updating and deleting a Role
type RoleService struct {
	Datastorer diygoapi.Datastorer
}

// Create is used to create a Role
func (s *RoleService) Create(ctx context.Context, r *diygoapi.CreateRoleRequest, adt diygoapi.Audit) (response *diygoapi.RoleResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var rolePermissions []*diygoapi.Permission
	rolePermissions, err = findPermissions(ctx, tx, r.Permissions)
	if err != nil {
		return nil, err
	}

	role := diygoapi.Role{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Code:        r.Code,
		Description: r.Description,
		Active:      r.Active,
		Permissions: rolePermissions,
	}

	err = createRoleTx(ctx, tx, role, adt)
	if err != nil {
		return nil, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	response = &diygoapi.RoleResponse{
		ExternalID:  role.ExternalID.String(),
		Code:        role.Code,
		Description: role.Description,
		Active:      role.Active,
		Permissions: role.Permissions,
	}

	return response, nil
}

// createRoleTx creates the role in the database
func createRoleTx(ctx context.Context, tx pgx.Tx, role diygoapi.Role, adt diygoapi.Audit) (err error) {
	err = role.Validate()
	if err != nil {
		return err
	}

	arg := datastore.CreateRoleParams{
		RoleID:          role.ID,
		RoleExtlID:      role.ExternalID.String(),
		RoleCd:          role.Code,
		Active:          role.Active,
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).CreateRole(ctx, arg)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	// should only impact exactly one record
	if rowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("Create() should insert 1 row, actual: %d", rowsAffected))
	}

	err = UpdateRolePermissions(ctx, tx, UpdateRolePermissionsParams{Role: role, Audit: adt})
	if err != nil {
		return err
	}

	return nil
}

// UpdateRolePermissionsParams is the parameters for the UpdateRolePermissions function
type UpdateRolePermissionsParams struct {
	Role  diygoapi.Role
	Audit diygoapi.Audit
}

// UpdateRolePermissions writes the Permissions attached to the role to the database.
// If there are existing permissions, in the database, they are removed.
func UpdateRolePermissions(ctx context.Context, tx pgx.Tx, params UpdateRolePermissionsParams) (err error) {
	_, err = datastore.New(tx).DeleteAllPermissions4Role(ctx, params.Role.ID)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	for _, rp := range params.Role.Permissions {
		createRolePermissionParams := datastore.CreateRolePermissionParams{
			RoleID:          params.Role.ID,
			PermissionID:    rp.ID,
			CreateAppID:     params.Audit.App.ID,
			CreateUserID:    params.Audit.User.NullUUID(),
			CreateTimestamp: params.Audit.Moment,
			UpdateAppID:     params.Audit.App.ID,
			UpdateUserID:    params.Audit.User.NullUUID(),
			UpdateTimestamp: params.Audit.Moment,
		}

		var rowsAffected int64
		rowsAffected, err = datastore.New(tx).CreateRolePermission(ctx, createRolePermissionParams)
		if err != nil {
			return errs.E(errs.Database, err)
		}

		// should only impact exactly one record
		if rowsAffected != 1 {
			return errs.E(errs.Database, fmt.Sprintf("Create() should insert 1 row, actual: %d", rowsAffected))
		}
	}

	return nil
}

// FindRoleByCode returns a Role and its permissions.
func FindRoleByCode(ctx context.Context, tx datastore.DBTX, code string) (diygoapi.Role, error) {
	dbRole, err := datastore.New(tx).FindRoleByCode(ctx, code)
	if err != nil {
		return diygoapi.Role{}, errs.E(errs.Database, err)
	}

	var dbPermissions []datastore.Permission
	dbPermissions, err = datastore.New(tx).FindRolePermissionsByRoleID(ctx, dbRole.RoleID)
	if err != nil {
		return diygoapi.Role{}, errs.E(errs.Database, err)
	}

	var permissions []*diygoapi.Permission
	if dbPermissions != nil {
		for _, dbp := range dbPermissions {
			p := &diygoapi.Permission{
				ID:          dbp.PermissionID,
				ExternalID:  secure.MustParseIdentifier(dbp.PermissionExtlID),
				Resource:    dbp.Resource,
				Operation:   dbp.Operation,
				Description: dbp.PermissionDescription,
				Active:      dbp.Active,
			}
			permissions = append(permissions, p)
		}
	}

	role := diygoapi.Role{
		ID:          dbRole.RoleID,
		ExternalID:  secure.MustParseIdentifier(dbRole.RoleExtlID),
		Code:        dbRole.RoleCd,
		Description: dbRole.RoleDescription,
		Active:      dbRole.Active,
		Permissions: permissions,
	}

	return role, nil
}

type assignOrgRoleParams struct {
	Role  diygoapi.Role
	User  *diygoapi.User
	Org   *diygoapi.Org
	Audit diygoapi.Audit
}

// assignOrgRoles assigns a role to a user for a given org.
func assignOrgRole(ctx context.Context, tx pgx.Tx, p assignOrgRoleParams) (err error) {
	params := datastore.CreateUsersRoleParams{
		UserID:          p.User.ID,
		RoleID:          p.Role.ID,
		OrgID:           p.Org.ID,
		CreateAppID:     p.Audit.App.ID,
		CreateUserID:    p.Audit.User.NullUUID(),
		CreateTimestamp: p.Audit.Moment,
		UpdateAppID:     p.Audit.App.ID,
		UpdateUserID:    p.Audit.User.NullUUID(),
		UpdateTimestamp: p.Audit.Moment,
	}

	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).CreateUsersRole(ctx, params)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	// should only impact exactly one record
	if rowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("CreateUsersRole() should insert 1 row, actual: %d", rowsAffected))
	}

	return nil
}

// findPermissions finds a list of permissions in the database using
// the Permission External ID first and if not given, the resource and
// operation.
func findPermissions(ctx context.Context, tx pgx.Tx, prs []*diygoapi.FindPermissionRequest) (aps []*diygoapi.Permission, err error) {

	// it's fine for zero permissions to be added as part of a role
	if len(prs) == 0 {
		return nil, nil
	}

	// if permissions are set as part of role create, find them in the db depending on
	// which key is sent (external id or resource/operation)
	for _, pr := range prs {
		var ap datastore.Permission
		if pr.ExternalID != "" {
			ap, err = datastore.New(tx).FindPermissionByExternalID(ctx, pr.ExternalID)
			if err != nil {
				return nil, errs.E(errs.Database, err)
			}
			aps = append(aps, newPermission(ap))
		} else {
			ap, err = datastore.New(tx).FindPermissionByResourceOperation(ctx, datastore.FindPermissionByResourceOperationParams{Resource: pr.Resource, Operation: pr.Operation})
			if err != nil {
				return nil, errs.E(errs.Database, err)
			}
			aps = append(aps, newPermission(ap))
		}
	}

	return aps, nil
}
