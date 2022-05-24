package userstore

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diy-go-api/datastore/datastoretest"
	"github.com/google/uuid"
)

func TestQueries_FindUserByUsername(t *testing.T) {
	t.Run("no row returned", func(t *testing.T) {
		c := qt.New(t)
		ds, _ := datastoretest.NewDatastore(t)
		//c.Cleanup(cleanup)
		ctx := context.Background()
		tx, err := ds.Pool().Begin(ctx)
		if err != nil {
			c.Fatal(err)
		}
		defer tx.Rollback(ctx)

		params := FindUserByUsernameParams{
			Username: "not there",
			OrgID:    uuid.New(),
		}

		_, err = New(tx).FindUserByUsername(ctx, params)
		c.Check(err, qt.ErrorIs, pgx.ErrNoRows)
	})
}
