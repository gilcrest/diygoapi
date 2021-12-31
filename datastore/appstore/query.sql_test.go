package appstore

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
)

func TestQueries_FindAppAPIKeysByAppExtlID(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)
		ds, _ := datastoretest.NewDatastore(t)

		q := New(ds.Pool())

		a, err := q.FindAppAPIKeysByAppExtlID(context.Background(), "booger")
		c.Assert(err, qt.IsNil)
		for _, row := range a {
			c.Log(row)
		}
		c.Assert(len(a) >= 1, qt.IsTrue)
		//c.Assert("a", qt.Equals, "Cyven2xyew89tIMcUEhf")
	})
}

func TestQueries_CreateApp(t *testing.T) {
	t.Run("numbers", func(t *testing.T) {
		c := qt.New(t)
		ds, _ := datastoretest.NewDatastore(t)
		//c.Cleanup(cleanup)
		ctx := context.Background()
		tx, err := ds.Pool().Begin(ctx)
		if err != nil {
			c.Fatal(err)
		}
		defer tx.Rollback(ctx)

		q := New(tx)

		cmd, err := q.CreateApp(ctx, CreateAppParams{})
		c.Assert(err, qt.IsNil)
		c.Assert(cmd.RowsAffected() > 0, qt.IsTrue)

		//c.Assert(len(a) >= 1, qt.IsTrue)
		//c.Assert("a", qt.Equals, "Cyven2xyew89tIMcUEhf")
	})
}
