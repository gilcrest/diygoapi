package service_test

import (
	"context"
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
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/service"
)

func TestOrgService(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		s := service.OrgService{
			Datastorer: ds,
		}
		r := service.CreateOrgRequest{
			Name:        "Test Org",
			Description: "Test Org created via TestCreateOrgService_Create",
			Kind:        "test",
		}

		ctx := context.Background()

		adt := findGenesisTestAudit(ctx, t, ds)

		got, err := s.Create(context.Background(), &r, adt)
		want := service.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                "Test Org",
			KindExternalID:      "test",
			Description:         "Test Org created via TestCreateOrgService_Create",
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
	t.Run("update", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		var (
			testOrg orgstore.FindOrgByNameRow
			err     error
		)
		testOrg, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, "Test Org")
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: ds,
		}
		r := service.UpdateOrgRequest{
			ExternalID:  testOrg.OrgExtlID,
			Name:        "Updated Test Org",
			Description: "Test Org updated via TestCreateOrgService_Update",
		}

		adt := findGenesisTestAudit(ctx, t, ds)

		var got service.OrgResponse
		got, err = s.Update(context.Background(), &r, adt)
		want := service.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                "Updated Test Org",
			KindExternalID:      "test",
			Description:         "Test Org updated via TestCreateOrgService_Update",
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
		testOrg, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, "Updated Test Org")
		if err != nil {
			t.Fatalf("FindOrgByName() error = %v", err)
		}

		s := service.OrgService{
			Datastorer: ds,
		}

		adt := findGenesisTestAudit(ctx, t, ds)

		var got service.OrgResponse
		got, err = s.FindByExternalID(context.Background(), testOrg.OrgExtlID)
		want := service.OrgResponse{
			ExternalID:          got.ExternalID,
			Name:                "Updated Test Org",
			KindExternalID:      "test",
			Description:         "Test Org updated via TestCreateOrgService_Update",
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
		testOrg, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, "Updated Test Org")
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

func findGenesisTestAudit(ctx context.Context, t *testing.T, ds datastore.Datastore) audit.Audit {
	t.Helper()

	var (
		findOrgByNameRow orgstore.FindOrgByNameRow
		err              error
	)
	findOrgByNameRow, err = orgstore.New(ds.Pool()).FindOrgByName(ctx, "genesis")
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
		AppName: "WOPR",
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
		Username: "pcollins",
		OrgID:    genesisOrg.ID,
	}

	var philCollinsDBUserRow userstore.FindUserByUsernameRow
	philCollinsDBUserRow, err = userstore.New(ds.Pool()).FindUserByUsername(ctx, findUserByUsernameParams)
	if err != nil {
		t.Fatalf("FindUserByUsername() error = %v", err)
	}

	genesisTestUser := user.User{
		ID:       philCollinsDBUserRow.UserID,
		Username: philCollinsDBUserRow.Username,
		Org:      genesisOrg,
		Profile: person.Profile{
			FirstName: philCollinsDBUserRow.FirstName,
			LastName:  philCollinsDBUserRow.LastName,
		},
	}

	adt := audit.Audit{
		App:    genesisApp,
		User:   genesisTestUser,
		Moment: time.Now(),
	}

	return adt
}
