package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diy-go-api"
	"github.com/gilcrest/diy-go-api/errs"
	"github.com/gilcrest/diy-go-api/secure"
	"github.com/gilcrest/diy-go-api/service"
	"github.com/gilcrest/diy-go-api/sqldb/datastore"
	"github.com/gilcrest/diy-go-api/sqldb/sqldbtest"
)

const (
	testOrgServiceOrgName               = "TestCreateOrgService_Create"
	testOrgServiceOrgDescription        = "Test Org created via TestCreateOrgService_Create"
	testOrgServiceOrgKind               = "test"
	testOrgServiceUpdatedOrgName        = "TestCreateOrgService_Update"
	testOrgServiceUpdatedOrgDescription = "Test Org updated via TestCreateOrgService_Update"
)

func TestOrgService(t *testing.T) {
	t.Run("create no request error", func(t *testing.T) {
		c := qt.New(t)

		eks := os.Getenv("ENCRYPT_KEY")
		if eks == "" {
			t.Fatal("no encryption key found")
		}

		// decode and retrieve encryption key
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.ParseEncryptionKey(eks)
		if err != nil {
			t.Fatal("secure.ParseEncryptionKey() error")
		}

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		s := service.OrgService{
			Datastorer:      db,
			APIKeyGenerator: secure.RandomGenerator{},
			EncryptionKey:   ek,
		}
		adt := findPrincipalTestAudit(ctx, c, tx)

		var got *diy.OrgResponse
		got, err = s.Create(context.Background(), nil, adt)
		c.Assert(errs.KindIs(errs.Validation, err), qt.IsTrue)
		c.Assert(err.Error(), qt.Equals, "CreateOrgRequest must have a value when creating an Org")
		c.Assert(got, qt.IsNil)
	})
	t.Run("create (with app)", func(t *testing.T) {
		c := qt.New(t)

		eks := os.Getenv("ENCRYPT_KEY")
		if eks == "" {
			t.Fatal("no encryption key found")
		}

		// decode and retrieve encryption key
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.ParseEncryptionKey(eks)
		if err != nil {
			t.Fatal("secure.ParseEncryptionKey() error")
		}

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		s := service.OrgService{
			Datastorer:      db,
			APIKeyGenerator: secure.RandomGenerator{},
			EncryptionKey:   ek,
		}
		r := diy.CreateOrgRequest{
			Name:        testOrgServiceOrgName + "_withApp",
			Description: testOrgServiceOrgDescription + "_withApp",
			Kind:        testOrgServiceOrgKind,
			CreateAppRequest: &diy.CreateAppRequest{
				Name:        testAppServiceAppName,
				Description: testAppServiceAppDescription,
			},
		}

		adt := findPrincipalTestAudit(ctx, c, tx)

		var got *diy.OrgResponse
		got, err = s.Create(context.Background(), &r, adt)
		c.Assert(err, qt.IsNil)
		want := &diy.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                testOrgServiceOrgName + "_withApp",
			KindExternalID:      testOrgServiceOrgKind,
			Description:         testOrgServiceOrgDescription + "_withApp",
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUserFirstName: adt.User.FirstName,
			CreateUserLastName:  adt.User.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUserFirstName: adt.User.FirstName,
			UpdateUserLastName:  adt.User.LastName,
			App: &diy.AppResponse{
				ExternalID:          got.App.ExternalID,
				Name:                testAppServiceAppName,
				Description:         testAppServiceAppDescription,
				CreateAppExtlID:     adt.App.ExternalID.String(),
				CreateUserFirstName: adt.User.FirstName,
				CreateUserLastName:  adt.User.LastName,
				UpdateAppExtlID:     adt.App.ExternalID.String(),
				UpdateUserFirstName: adt.User.FirstName,
				UpdateUserLastName:  adt.User.LastName,
				APIKeys:             nil,
			},
		}
		ignoreFields := []string{"ExternalID", "CreateDateTime", "UpdateDateTime", "App.CreateDateTime", "App.UpdateDateTime", "App.APIKeys"}
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(diy.OrgResponse{}, ignoreFields...)), want)
	})
	t.Run("delete (with app)", func(t *testing.T) {
		c := qt.New(t)

		var (
			testOrg datastore.FindOrgByNameRow
			err     error
		)

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		testOrg, err = datastore.New(tx).FindOrgByName(ctx, testOrgServiceOrgName+"_withApp")
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: db,
		}

		var got diy.DeleteResponse
		got, err = s.Delete(context.Background(), testOrg.OrgExtlID)
		want := diy.DeleteResponse{
			ExternalID: testOrg.OrgExtlID,
			Deleted:    true,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(), want)
	})
	t.Run("update", func(t *testing.T) {
		c := qt.New(t)

		var (
			testOrg datastore.FindOrgByNameRow
			err     error
		)

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		testOrg, err = datastore.New(tx).FindOrgByName(ctx, testOrgServiceOrgName)
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: db,
		}
		r := diy.UpdateOrgRequest{
			ExternalID:  testOrg.OrgExtlID,
			Name:        testOrgServiceUpdatedOrgName,
			Description: testOrgServiceUpdatedOrgDescription,
		}

		adt := findPrincipalTestAudit(ctx, c, tx)

		var got *diy.OrgResponse
		got, err = s.Update(context.Background(), &r, adt)
		want := &diy.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                testOrgServiceUpdatedOrgName,
			KindExternalID:      testOrgServiceOrgKind,
			Description:         testOrgServiceUpdatedOrgDescription,
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUserFirstName: adt.User.FirstName,
			CreateUserLastName:  adt.User.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUserFirstName: adt.User.FirstName,
			UpdateUserLastName:  adt.User.LastName,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(diy.OrgResponse{}, "CreateDateTime", "UpdateDateTime")), want)
	})
	t.Run("findByExtlID", func(t *testing.T) {
		c := qt.New(t)

		var (
			testOrg datastore.FindOrgByNameRow
			err     error
		)

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		testOrg, err = datastore.New(tx).FindOrgByName(ctx, testOrgServiceUpdatedOrgName)
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: db,
		}

		adt := findPrincipalTestAudit(ctx, c, tx)

		var got *diy.OrgResponse
		got, err = s.FindByExternalID(context.Background(), testOrg.OrgExtlID)
		want := &diy.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                testOrgServiceUpdatedOrgName,
			KindExternalID:      testOrgServiceOrgKind,
			Description:         testOrgServiceUpdatedOrgDescription,
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUserFirstName: adt.User.FirstName,
			CreateUserLastName:  adt.User.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUserFirstName: adt.User.FirstName,
			UpdateUserLastName:  adt.User.LastName,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(diy.OrgResponse{}, "CreateDateTime", "UpdateDateTime")), want)
	})
	t.Run("findAll", func(t *testing.T) {
		c := qt.New(t)

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		s := service.OrgService{
			Datastorer: db,
		}

		var (
			got []*diy.OrgResponse
			err error
		)
		got, err = s.FindAll(ctx)
		c.Assert(err, qt.IsNil)
		c.Assert(len(got) >= 1, qt.IsTrue, qt.Commentf("orgs found = %d", len(got)))
		c.Logf("orgs found = %d", len(got))
	})
	t.Run("delete", func(t *testing.T) {
		c := qt.New(t)

		var (
			testOrg datastore.FindOrgByNameRow
			err     error
		)

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		testOrg, err = datastore.New(tx).FindOrgByName(ctx, testOrgServiceUpdatedOrgName)
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: db,
		}

		var got diy.DeleteResponse
		got, err = s.Delete(context.Background(), testOrg.OrgExtlID)
		want := diy.DeleteResponse{
			ExternalID: testOrg.OrgExtlID,
			Deleted:    true,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(), want)
	})
}

// findPrincipalTestAudit returns a diy.Audit with the Principal Org, App and a Test User
func findPrincipalTestAudit(ctx context.Context, c *qt.C, tx pgx.Tx) diy.Audit {
	c.Helper()

	var (
		findOrgByNameRow datastore.FindOrgByNameRow
		err              error
	)
	findOrgByNameRow, err = datastore.New(tx).FindOrgByName(ctx, service.PrincipalOrgName)
	if err != nil {
		c.Fatalf("FindOrgByName() error = %v", err)
	}

	k := &diy.OrgKind{
		ID:          findOrgByNameRow.OrgKindID,
		ExternalID:  findOrgByNameRow.OrgKindExtlID,
		Description: findOrgByNameRow.OrgKindDesc,
	}

	genesisOrg := &diy.Org{
		ID:          findOrgByNameRow.OrgID,
		ExternalID:  secure.MustParseIdentifier(findOrgByNameRow.OrgExtlID),
		Name:        findOrgByNameRow.OrgName,
		Description: findOrgByNameRow.OrgDescription,
		Kind:        k,
	}

	findAppByNameParams := datastore.FindAppByNameParams{
		OrgID:   findOrgByNameRow.OrgID,
		AppName: service.PrincipalAppName,
	}

	var genesisDBAppRow datastore.FindAppByNameRow
	genesisDBAppRow, err = datastore.New(tx).FindAppByName(context.Background(), findAppByNameParams)
	if err != nil {
		c.Fatalf("FindTestApp() error = %v", err)
	}

	genesisApp := &diy.App{
		ID:          genesisDBAppRow.AppID,
		ExternalID:  secure.MustParseIdentifier(genesisDBAppRow.AppExtlID),
		Org:         genesisOrg,
		Name:        genesisDBAppRow.AppName,
		Description: genesisDBAppRow.AppDescription,
		APIKeys:     nil,
	}

	var testOrg *diy.Org
	testOrg, err = service.FindOrgByName(ctx, tx, service.TestOrgName)
	if err != nil {
		c.Fatalf("FindOrgByName() error = %v", err)
	}

	var testRole diy.Role
	testRole, err = service.FindRoleByCode(ctx, tx, service.TestRoleCode)
	if err != nil {
		c.Fatalf("FindRoleByCode() error = %v", err)
	}

	findUsersByOrgRoleParams := datastore.FindUsersByOrgRoleParams{
		OrgID:  testOrg.ID,
		RoleID: testRole.ID,
	}

	var usersRole []datastore.UsersRole
	usersRole, err = datastore.New(tx).FindUsersByOrgRole(ctx, findUsersByOrgRoleParams)
	if err != nil {
		c.Fatalf("FindUsersByOrgRole() error = %v", err)
	}

	var u *diy.User
	for i, ur := range usersRole {
		u, err = service.FindUserByID(ctx, tx, ur.UserID)
		if err != nil {
			c.Fatalf("FindUserByID() error = %v", err)
		}
		// theoretically, only one user should have this role, so we
		// don't need this break, but just in case...
		if i == 0 {
			break
		}
	}

	adt := diy.Audit{
		App:    genesisApp,
		User:   u,
		Moment: time.Now(),
	}

	return adt
}
