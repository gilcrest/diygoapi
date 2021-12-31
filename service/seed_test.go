package service_test

import (
	"context"
	"os"
	"testing"

	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/service"

	qt "github.com/frankban/quicktest"
)

func TestSeedService_Seed(t *testing.T) {
	t.Skip()
	t.Run("numbers", func(t *testing.T) {
		c := qt.New(t)

		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		const keyEnv string = "ENCRYPT_KEY"
		ekey, ok := os.LookupEnv(keyEnv)
		if !ok {
			c.Fatalf("%s not set\n", keyEnv)
		}
		ek, err := secure.ParseEncryptionKey(ekey)
		if err != nil {
			c.Fatal(err)
		}

		sr := service.SeedService{
			Datastorer:            ds,
			CryptoRandomGenerator: secure.CryptoRandomGenerator{},
			EncryptionKey:         ek,
		}
		r := service.SeedRequest{
			OrgName:           "WOPR",
			OrgDescription:    "Seed Org",
			AppName:           "Seed",
			AppDescription:    "Seed App",
			SeedUsername:      "dan@dangillis.dev",
			SeedUserFirstName: "Dan",
			SeedUserLastName:  "Gillis",
		}

		_, err = sr.Seed(context.Background(), &r)
		c.Assert(err, qt.IsNil)
	})
}
