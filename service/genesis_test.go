package service_test

import (
	"context"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/domain/secure/random"
	"github.com/gilcrest/go-api-basic/service"
)

func TestGenesisService_Seed(t *testing.T) {
	t.Skip()
	t.Run("standard", func(t *testing.T) {
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

		sr := service.GenesisService{
			Datastorer:            ds,
			RandomStringGenerator: random.CryptoGenerator{},
			EncryptionKey:         ek,
		}
		r := service.GenesisRequest{
			SeedUsername:      "dan@dangillis.dev",
			SeedUserFirstName: "Dan",
			SeedUserLastName:  "Gillis",
		}

		_, err = sr.Seed(context.Background(), &r)
		c.Assert(err, qt.IsNil)
	})
}
