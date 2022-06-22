package service_test

import (
	"context"
	"os"
	"strings"
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
	testOrgServiceOrgName               = "TestCreateOrgService_Create"
	testOrgServiceOrgDescription        = "Test Org created via TestCreateOrgService_Create"
	testOrgServiceOrgKind               = "test"
	testOrgServiceUpdatedOrgName        = "TestCreateOrgService_Update"
	testOrgServiceUpdatedOrgDescription = "Test Org updated via TestCreateOrgService_Update"
)

func TestOrgService(t *testing.T) {
	t.Run("create (without app)", func(t *testing.T) {
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

		s := service.CreateOrgService{
			Datastorer:            ds,
			RandomStringGenerator: random.CryptoGenerator{},
			EncryptionKey:         ek,
		}
		r := service.CreateOrgRequest{
			Name:        testOrgServiceOrgName,
			Description: testOrgServiceOrgDescription,
			Kind:        testOrgServiceOrgKind,
		}

		ctx := context.Background()

		adt := findPrincipalTestAudit(ctx, t, ds)

		var got service.OrgResponse
		got, err = s.Create(context.Background(), &r, adt)
		want := service.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                testOrgServiceOrgName,
			KindExternalID:      testOrgServiceOrgKind,
			Description:         testOrgServiceOrgDescription,
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
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(service.OrgResponse{}, "CreateDateTime", "UpdateDateTime")), want)
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

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		s := service.CreateOrgService{
			Datastorer:            ds,
			RandomStringGenerator: random.CryptoGenerator{},
			EncryptionKey:         ek,
		}
		r := service.CreateOrgRequest{
			Name:        testOrgServiceOrgName + "_withApp",
			Description: testOrgServiceOrgDescription + "_withApp",
			Kind:        testOrgServiceOrgKind,
			App: service.CreateAppRequest{
				Name:        testAppServiceAppName,
				Description: testAppServiceAppDescription,
			},
		}

		ctx := context.Background()

		adt := findPrincipalTestAudit(ctx, t, ds)

		var got service.OrgResponse
		got, err = s.Create(context.Background(), &r, adt)
		want := service.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                testOrgServiceOrgName + "_withApp",
			KindExternalID:      testOrgServiceOrgKind,
			Description:         testOrgServiceOrgDescription + "_withApp",
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUsername:      adt.User.Username,
			CreateUserFirstName: adt.User.Profile.FirstName,
			CreateUserLastName:  adt.User.Profile.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUsername:      adt.User.Username,
			UpdateUserFirstName: adt.User.Profile.FirstName,
			UpdateUserLastName:  adt.User.Profile.LastName,
			App: service.AppResponse{
				ExternalID:          got.App.ExternalID,
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
				APIKeys:             nil,
			},
		}
		c.Assert(err, qt.IsNil)
		ignoreFields := []string{"ExternalID", "CreateDateTime", "UpdateDateTime", "App.CreateDateTime", "App.UpdateDateTime", "App.APIKeys"}
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(service.OrgResponse{}, ignoreFields...)), want)
	})
	t.Run("delete (with app)", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		var (
			testOrg orgstore.FindOrgByNameRow
			err     error
		)
		testOrg, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, testOrgServiceOrgName+"_withApp")
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: ds,
		}

		var got service.DeleteResponse
		got, err = s.Delete(context.Background(), testOrg.OrgExtlID)
		want := service.DeleteResponse{
			ExternalID: testOrg.OrgExtlID,
			Deleted:    true,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(), want)
	})
	t.Run("update", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		var (
			testOrg orgstore.FindOrgByNameRow
			err     error
		)
		testOrg, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, testOrgServiceOrgName)
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: ds,
		}
		r := service.UpdateOrgRequest{
			ExternalID:  testOrg.OrgExtlID,
			Name:        testOrgServiceUpdatedOrgName,
			Description: testOrgServiceUpdatedOrgDescription,
		}

		adt := findPrincipalTestAudit(ctx, t, ds)

		var got service.OrgResponse
		got, err = s.Update(context.Background(), &r, adt)
		want := service.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                testOrgServiceUpdatedOrgName,
			KindExternalID:      testOrgServiceOrgKind,
			Description:         testOrgServiceUpdatedOrgDescription,
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
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(service.OrgResponse{}, "CreateDateTime", "UpdateDateTime")), want)
	})
	t.Run("findByExtlID", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		var (
			testOrg orgstore.FindOrgByNameRow
			err     error
		)
		testOrg, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, testOrgServiceUpdatedOrgName)
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: ds,
		}

		adt := findPrincipalTestAudit(ctx, t, ds)

		var got service.OrgResponse
		got, err = s.FindByExternalID(context.Background(), testOrg.OrgExtlID)
		want := service.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                testOrgServiceUpdatedOrgName,
			KindExternalID:      testOrgServiceOrgKind,
			Description:         testOrgServiceUpdatedOrgDescription,
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
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(service.OrgResponse{}, "CreateDateTime", "UpdateDateTime")), want)
	})
	t.Run("findAll", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		s := service.OrgService{
			Datastorer: ds,
		}

		var (
			got []service.OrgResponse
			err error
		)
		got, err = s.FindAll(ctx)
		c.Assert(err, qt.IsNil)
		c.Assert(len(got) >= 1, qt.IsTrue, qt.Commentf("orgs found = %d", len(got)))
		c.Logf("orgs found = %d", len(got))
	})
	t.Run("delete", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		var (
			testOrg orgstore.FindOrgByNameRow
			err     error
		)
		testOrg, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, testOrgServiceUpdatedOrgName)
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: ds,
		}

		var got service.DeleteResponse
		got, err = s.Delete(context.Background(), testOrg.OrgExtlID)
		want := service.DeleteResponse{
			ExternalID: testOrg.OrgExtlID,
			Deleted:    true,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(), want)
	})
}

// findPrincipalTestAudit returns an audit.Audit with the Principal Org, App and a Test User
func findPrincipalTestAudit(ctx context.Context, t *testing.T, ds datastore.Datastore) audit.Audit {
	t.Helper()

	var (
		findOrgByNameRow orgstore.FindOrgByNameRow
		err              error
	)
	findOrgByNameRow, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, service.PrincipalOrgName)
	if err != nil {
		t.Fatalf("FindOrgByName() error = %v", err)
	}

	k := org.Kind{
		ID:          findOrgByNameRow.OrgKindID,
		ExternalID:  findOrgByNameRow.OrgKindExtlID,
		Description: findOrgByNameRow.OrgKindDesc,
	}

	genesisOrg := org.Org{
		ID:          findOrgByNameRow.OrgID,
		ExternalID:  secure.MustParseIdentifier(findOrgByNameRow.OrgExtlID),
		Name:        findOrgByNameRow.OrgName,
		Description: findOrgByNameRow.OrgDescription,
		Kind:        k,
	}

	findAppByNameParams := appstore.FindAppByNameParams{
		OrgID:   findOrgByNameRow.OrgID,
		AppName: service.PrincipalAppName,
	}

	var genesisDBAppRow appstore.FindAppByNameRow
	genesisDBAppRow, err = appstore.New(ds.Pool()).FindAppByName(context.Background(), findAppByNameParams)
	if err != nil {
		t.Fatalf("FindTestApp() error = %v", err)
	}

	genesisApp := app.App{
		ID:          genesisDBAppRow.AppID,
		ExternalID:  secure.MustParseIdentifier(genesisDBAppRow.AppExtlID),
		Org:         genesisOrg,
		Name:        genesisDBAppRow.AppName,
		Description: genesisDBAppRow.AppDescription,
		APIKeys:     nil,
	}

	findUserByUsernameParams := userstore.FindUserByUsernameParams{
		Username: strings.TrimSpace(service.PrincipalTestUsername),
		OrgID:    genesisOrg.ID,
	}

	var row userstore.FindUserByUsernameRow
	row, err = userstore.New(ds.Pool()).FindUserByUsername(ctx, findUserByUsernameParams)
	if err != nil {
		t.Fatalf("FindUserByUsername() error = %v", err)
	}

	genesisTestUser := user.User{
		ID:       row.UserID,
		Username: row.Username,
		Org:      genesisOrg,
		Profile: person.Profile{
			FirstName: row.FirstName,
			LastName:  row.LastName,
		},
	}

	adt := audit.Audit{
		App:    genesisApp,
		User:   genesisTestUser,
		Moment: time.Now(),
	}

	return adt
}
