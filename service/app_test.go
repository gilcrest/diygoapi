package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/gilcrest/diy-go-api/datastore"
	"github.com/gilcrest/diy-go-api/datastore/appstore"
	"github.com/gilcrest/diy-go-api/datastore/datastoretest"
	"github.com/gilcrest/diy-go-api/datastore/orgstore"
	"github.com/gilcrest/diy-go-api/datastore/userstore"
	"github.com/gilcrest/diy-go-api/domain/app"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/person"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/gilcrest/diy-go-api/domain/secure/random"
	"github.com/gilcrest/diy-go-api/domain/user"
	"github.com/gilcrest/diy-go-api/service"
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

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		s := service.AppService{
			Datastorer:            ds,
			RandomStringGenerator: random.CryptoGenerator{},
			EncryptionKey:         ek,
		}
		r := service.CreateAppRequest{
			Name:        testAppServiceAppName,
			Description: testAppServiceAppDescription,
		}

		ctx := context.Background()

		adt := findTestAudit(ctx, t, ds)

		var got service.AppResponse
		got, err = s.Create(context.Background(), &r, adt)
		want := service.AppResponse{
			Name:                testAppServiceAppName,
			Description:         testAppServiceAppDescription,
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUsername:      adt.User.Username,
			CreateUserFirstName: adt.User.Profile.FirstName,
			CreateUserLastName:  adt.User.Profile.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUsername:      adt.User.Username,
			UpdateUserFirstName: adt.User.Profile.FirstName,
			UpdateUserLastName:  adt.User.Profile.LastName,
		}
		ignoreFields := []string{"ExternalID", "CreateDateTime", "UpdateDateTime", "APIKeys"}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(service.AppResponse{}, ignoreFields...)), want)
	})
	t.Run("update", func(t *testing.T) {
		c := qt.New(t)

		var err error

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()
		adt := findTestAudit(ctx, t, ds)

		findAppByNameParams := appstore.FindAppByNameParams{
			OrgID:   adt.App.Org.ID,
			AppName: testAppServiceAppName,
		}

		var testAppRow appstore.FindAppByNameRow
		testAppRow, err = appstore.New(ds.Pool()).FindAppByName(ctx, findAppByNameParams)
		if err != nil {
			t.Fatalf("FindAppByName() error = %v", err)
		}

		s := service.AppService{
			Datastorer: ds,
		}
		r := service.UpdateAppRequest{
			ExternalID:  testAppRow.AppExtlID,
			Name:        testAppServiceUpdatedAppName,
			Description: testAppServiceUpdatedAppDescription,
		}

		var got service.AppResponse
		got, err = s.Update(context.Background(), &r, adt)
		want := service.AppResponse{
			Name:                testAppServiceUpdatedAppName,
			Description:         testAppServiceUpdatedAppDescription,
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUsername:      adt.User.Username,
			CreateUserFirstName: adt.User.Profile.FirstName,
			CreateUserLastName:  adt.User.Profile.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUsername:      adt.User.Username,
			UpdateUserFirstName: adt.User.Profile.FirstName,
			UpdateUserLastName:  adt.User.Profile.LastName,
		}
		ignoreFields := []string{"ExternalID", "CreateDateTime", "UpdateDateTime", "APIKeys"}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(service.AppResponse{}, ignoreFields...)), want)
	})
	t.Run("findByExtlID", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()
		adt := findTestAudit(ctx, t, ds)

		findAppByNameParams := appstore.FindAppByNameParams{
			OrgID:   adt.App.Org.ID,
			AppName: testAppServiceUpdatedAppName,
		}

		var (
			testAppRow appstore.FindAppByNameRow
			err        error
		)
		testAppRow, err = appstore.New(ds.Pool()).FindAppByName(ctx, findAppByNameParams)
		if err != nil {
			t.Fatalf("FindAppByName() error = %v", err)
		}

		s := service.AppService{
			Datastorer: ds,
		}

		var got service.AppResponse
		got, err = s.FindByExternalID(context.Background(), testAppRow.AppExtlID)
		want := service.AppResponse{
			ExternalID:          got.ExternalID,
			Name:                testAppServiceUpdatedAppName,
			Description:         testAppServiceUpdatedAppDescription,
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUsername:      adt.User.Username,
			CreateUserFirstName: adt.User.Profile.FirstName,
			CreateUserLastName:  adt.User.Profile.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUsername:      adt.User.Username,
			UpdateUserFirstName: adt.User.Profile.FirstName,
			UpdateUserLastName:  adt.User.Profile.LastName,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(service.AppResponse{}, "CreateDateTime", "UpdateDateTime")), want)
	})
	t.Run("findAll", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		s := service.AppService{
			Datastorer: ds,
		}

		var (
			got []service.AppResponse
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

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()
		adt := findTestAudit(ctx, t, ds)

		findAppByNameParams := appstore.FindAppByNameParams{
			OrgID:   adt.App.Org.ID,
			AppName: testAppServiceUpdatedAppName,
		}

		var testAppRow appstore.FindAppByNameRow
		testAppRow, err = appstore.New(ds.Pool()).FindAppByName(ctx, findAppByNameParams)
		if err != nil {
			t.Fatalf("FindAppByName() error = %v", err)
		}

		s := service.AppService{
			Datastorer: ds,
		}

		var got service.DeleteResponse
		got, err = s.Delete(context.Background(), testAppRow.AppExtlID)
		want := service.DeleteResponse{
			ExternalID: testAppRow.AppExtlID,
			Deleted:    true,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.Equals, want)
	})
}

func findTestAudit(ctx context.Context, t *testing.T, ds datastore.Datastore) audit.Audit {
	t.Helper()

	var (
		findOrgByNameRow orgstore.FindOrgByNameRow
		err              error
	)
	findOrgByNameRow, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, service.TestOrgName)
	if err != nil {
		t.Fatalf("FindOrgByName() error = %v", err)
	}

	testOrg := org.Org{
		ID:          findOrgByNameRow.OrgID,
		ExternalID:  secure.MustParseIdentifier(findOrgByNameRow.OrgExtlID),
		Name:        findOrgByNameRow.OrgName,
		Description: findOrgByNameRow.OrgDescription,
		Kind: org.Kind{
			ID:          findOrgByNameRow.OrgKindID,
			ExternalID:  findOrgByNameRow.OrgKindExtlID,
			Description: findOrgByNameRow.OrgKindDesc,
		},
	}

	findAppByNameParams := appstore.FindAppByNameParams{
		OrgID:   findOrgByNameRow.OrgID,
		AppName: service.TestAppName,
	}

	var testDBAppRow appstore.FindAppByNameRow
	testDBAppRow, err = appstore.New(ds.Pool()).FindAppByName(context.Background(), findAppByNameParams)
	if err != nil {
		t.Fatalf("FindTestApp() error = %v", err)
	}

	testApp := app.App{
		ID:          testDBAppRow.AppID,
		ExternalID:  secure.MustParseIdentifier(testDBAppRow.AppExtlID),
		Org:         testOrg,
		Name:        testDBAppRow.AppName,
		Description: testDBAppRow.AppDescription,
		APIKeys:     nil,
	}

	findUserByUsernameParams := userstore.FindUserByUsernameParams{
		Username: service.TestUsername,
		OrgID:    testOrg.ID,
	}

	var findUserByUsernameRow userstore.FindUserByUsernameRow
	findUserByUsernameRow, err = userstore.New(ds.Pool()).FindUserByUsername(ctx, findUserByUsernameParams)
	if err != nil {
		t.Fatalf("FindUserByUsername() error = %v", err)
	}

	testUser := user.User{
		ID:       findUserByUsernameRow.UserID,
		Username: findUserByUsernameRow.Username,
		Org:      testOrg,
		Profile: person.Profile{
			FirstName: findUserByUsernameRow.FirstName,
			LastName:  findUserByUsernameRow.LastName,
		},
	}

	adt := audit.Audit{
		App:    testApp,
		User:   testUser,
		Moment: time.Now(),
	}

	return adt
}
