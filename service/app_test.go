package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/secure"
	"github.com/gilcrest/diygoapi/service"
	"github.com/gilcrest/diygoapi/sqldb/datastore"
	"github.com/gilcrest/diygoapi/sqldb/sqldbtest"
)

const (
	testAppServiceAppName               = "TestAppService_Create"
	testAppServiceAppDescription        = "Test App created via TestAppService_Create"
	testAppServiceUpdatedAppName        = "TestAppService_Update"
	testAppServiceUpdatedAppDescription = "Test App updated via TestAppService_Update"
)

func TestAppService(t *testing.T) {
	t.Run("create", func(t *testing.T) {
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
		var tx pgx.Tx
		ctx := context.Background()
		tx, err = db.BeginTx(ctx)
		if err != nil {
			c.Fatalf("BeginTx() error = %v", err)
		}
		c.Cleanup(func() { _ = db.RollbackTx(ctx, tx, err) })

		s := service.AppService{
			Datastorer:      db,
			APIKeyGenerator: secure.RandomGenerator{},
			EncryptionKey:   ek,
		}
		r := diygoapi.CreateAppRequest{
			Name:        testAppServiceAppName,
			Description: testAppServiceAppDescription,
		}

		adt := findTestAudit(ctx, c, tx)

		var got *diygoapi.AppResponse
		got, err = s.Create(context.Background(), &r, adt)
		want := &diygoapi.AppResponse{
			Name:                testAppServiceAppName,
			Description:         testAppServiceAppDescription,
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUserFirstName: adt.User.FirstName,
			CreateUserLastName:  adt.User.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUserFirstName: adt.User.FirstName,
			UpdateUserLastName:  adt.User.LastName,
		}
		ignoreFields := []string{"ExternalID", "CreateDateTime", "UpdateDateTime", "APIKeys"}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(diygoapi.AppResponse{}, ignoreFields...)), want)
	})
	t.Run("update", func(t *testing.T) {
		c := qt.New(t)

		var err error

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		var tx pgx.Tx
		ctx := context.Background()
		tx, err = db.BeginTx(ctx)
		if err != nil {
			c.Fatalf("BeginTx() error = %v", err)
		}
		c.Cleanup(func() { _ = db.RollbackTx(ctx, tx, err) })

		adt := findTestAudit(ctx, c, tx)

		findAppByNameParams := datastore.FindAppByNameParams{
			OrgID:   adt.App.Org.ID,
			AppName: testAppServiceAppName,
		}

		var testAppRow datastore.FindAppByNameRow
		testAppRow, err = datastore.New(tx).FindAppByName(ctx, findAppByNameParams)
		if err != nil {
			t.Fatalf("FindAppByName() error = %v", err)
		}

		s := service.AppService{
			Datastorer: db,
		}
		r := diygoapi.UpdateAppRequest{
			ExternalID:  testAppRow.AppExtlID,
			Name:        testAppServiceUpdatedAppName,
			Description: testAppServiceUpdatedAppDescription,
		}

		var got *diygoapi.AppResponse
		got, err = s.Update(context.Background(), &r, adt)
		want := &diygoapi.AppResponse{
			Name:                testAppServiceUpdatedAppName,
			Description:         testAppServiceUpdatedAppDescription,
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUserFirstName: adt.User.FirstName,
			CreateUserLastName:  adt.User.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUserFirstName: adt.User.FirstName,
			UpdateUserLastName:  adt.User.LastName,
		}
		ignoreFields := []string{"ExternalID", "CreateDateTime", "UpdateDateTime", "APIKeys"}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(diygoapi.AppResponse{}, ignoreFields...)), want)
	})
	t.Run("findByExtlID", func(t *testing.T) {
		c := qt.New(t)

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		tx, err := db.BeginTx(ctx)
		if err != nil {
			c.Fatalf("BeginTx() error = %v", err)
		}
		c.Cleanup(func() { _ = db.RollbackTx(ctx, tx, err) })

		adt := findTestAudit(ctx, c, tx)

		var testAppRow datastore.FindAppByNameRow
		findAppByNameParams := datastore.FindAppByNameParams{
			OrgID:   adt.App.Org.ID,
			AppName: testAppServiceUpdatedAppName,
		}

		testAppRow, err = datastore.New(tx).FindAppByName(ctx, findAppByNameParams)
		if err != nil {
			t.Fatalf("FindAppByName() error = %v", err)
		}

		s := service.AppService{
			Datastorer: db,
		}

		var got *diygoapi.AppResponse
		got, err = s.FindByExternalID(context.Background(), testAppRow.AppExtlID)
		want := &diygoapi.AppResponse{
			ExternalID:          got.ExternalID,
			Name:                testAppServiceUpdatedAppName,
			Description:         testAppServiceUpdatedAppDescription,
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUserFirstName: adt.User.FirstName,
			CreateUserLastName:  adt.User.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUserFirstName: adt.User.FirstName,
			UpdateUserLastName:  adt.User.LastName,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(diygoapi.AppResponse{}, "CreateDateTime", "UpdateDateTime")), want)
	})
	t.Run("findAll", func(t *testing.T) {
		c := qt.New(t)

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		s := service.AppService{
			Datastorer: db,
		}

		var (
			got []*diygoapi.AppResponse
			err error
		)
		got, err = s.FindAll(ctx)
		c.Assert(err, qt.IsNil)
		c.Assert(len(got) >= 1, qt.IsTrue, qt.Commentf("apps found = %d, should be at least 1", len(got)))
		c.Logf("apps found = %d", len(got))
	})
	t.Run("delete", func(t *testing.T) {
		c := qt.New(t)

		var err error

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		var tx pgx.Tx
		ctx := context.Background()
		tx, err = db.BeginTx(ctx)
		if err != nil {
			c.Fatalf("BeginTx() error = %v", err)
		}
		c.Cleanup(func() { _ = db.RollbackTx(ctx, tx, err) })

		adt := findTestAudit(ctx, c, tx)

		findAppByNameParams := datastore.FindAppByNameParams{
			OrgID:   adt.App.Org.ID,
			AppName: testAppServiceUpdatedAppName,
		}

		var testAppRow datastore.FindAppByNameRow
		testAppRow, err = datastore.New(tx).FindAppByName(ctx, findAppByNameParams)
		if err != nil {
			t.Fatalf("FindAppByName() error = %v", err)
		}

		s := service.AppService{
			Datastorer: db,
		}

		var got diygoapi.DeleteResponse
		got, err = s.Delete(context.Background(), testAppRow.AppExtlID)
		want := diygoapi.DeleteResponse{
			ExternalID: testAppRow.AppExtlID,
			Deleted:    true,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.Equals, want)
	})
}

func findTestAudit(ctx context.Context, c *qt.C, tx datastore.DBTX) diygoapi.Audit {
	c.Helper()

	var err error

	var testOrg *diygoapi.Org
	testOrg, err = service.FindOrgByName(ctx, tx, service.TestOrgName)
	if err != nil {
		c.Fatalf("FindOrgByName() error = %v", err)
	}

	var testApp *diygoapi.App
	testApp, err = service.FindAppByName(ctx, tx, testOrg, service.TestAppName)
	if err != nil {
		c.Fatalf("FindOrgByName() error = %v", err)
	}

	var testRole diygoapi.Role
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

	var u *diygoapi.User
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

	adt := diygoapi.Audit{
		App:    testApp,
		User:   u,
		Moment: time.Now(),
	}

	return adt
}
