package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/appstore"
	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/datastore/orgstore"
	"github.com/gilcrest/go-api-basic/datastore/userstore"
	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/domain/secure/random"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/service"
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
			Name:        "Test App",
			Description: "Test App created via TestAppService_Create",
		}

		ctx := context.Background()

		adt := findTestAudit(ctx, t, ds)

		var got service.AppResponse
		got, err = s.Create(context.Background(), &r, adt)
		want := service.AppResponse{
			Name:                "Test App",
			Description:         "Test App created via TestAppService_Create",
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
			AppName: "Test App",
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
			Name:        "Updated Test App",
			Description: "Test App updated via TestAppService_Update",
		}

		var got service.AppResponse
		got, err = s.Update(context.Background(), &r, adt)
		want := service.AppResponse{
			Name:                "Updated Test App",
			Description:         "Test App updated via TestAppService_Update",
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
}

func findTestAudit(ctx context.Context, t *testing.T, ds datastore.Datastore) audit.Audit {
	t.Helper()

	var (
		findOrgByNameRow orgstore.FindOrgByNameRow
		err              error
	)
	findOrgByNameRow, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, "test")
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
		AppName: "test",
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
		Username: "shackett",
		OrgID:    testOrg.ID,
	}

	var steveHackettDBUserRow userstore.FindUserByUsernameRow
	steveHackettDBUserRow, err = userstore.New(ds.Pool()).FindUserByUsername(ctx, findUserByUsernameParams)
	if err != nil {
		t.Fatalf("FindUserByUsername() error = %v", err)
	}

	testUser := user.User{
		ID:       steveHackettDBUserRow.UserID,
		Username: steveHackettDBUserRow.Username,
		Org:      testOrg,
		Profile: person.Profile{
			FirstName: steveHackettDBUserRow.FirstName,
			LastName:  steveHackettDBUserRow.LastName,
		},
	}

	adt := audit.Audit{
		App:    testApp,
		User:   testUser,
		Moment: time.Now(),
	}

	return adt
}
