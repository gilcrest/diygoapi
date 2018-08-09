package auth

import (
	"context"
	"reflect"
	"testing"

	"github.com/gilcrest/go-API-template/appuser"
	"github.com/gilcrest/go-API-template/env"
	"github.com/rs/zerolog"
)

func Test_Authorise(t *testing.T) {
	// Get an empty context with a Cancel function included
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel ctx as soon as test returns.

	// Initializes "environment" struct type
	ev, err := env.NewEnv(zerolog.DebugLevel)
	if err != nil {
		t.Errorf("Error Initializing env, err = %s", err)
	}

	// set the *sql.Tx for the
	// datastore within the environment
	err = ev.DS.SetTx(ctx, nil)
	if err != nil {
		t.Error(err)
	}

	usr := new(appuser.User)

	type args struct {
		ctx context.Context
		env *env.Env
		c   *Credentials
	}

	creds := new(Credentials)
	creds.Username = "asdf"
	creds.Password = "wutang#1"

	tests := []struct {
		name    string
		args    args
		want    *appuser.User
		wantErr bool
	}{
		{"Test Dan", args{ctx: ctx, env: ev, c: creds}, usr, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Authorise(tt.args.ctx, tt.args.env, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("Authorise() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Authorise() = %v, want %v", got, tt.want)
			}
		})
	}
}
