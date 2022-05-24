package audit

import (
	"net/http"
	"time"

	"github.com/gilcrest/diy-go-api/domain/app"
	"github.com/gilcrest/diy-go-api/domain/user"
)

// Audit represents the moment an app/user interacted with the system
type Audit struct {
	App    app.App
	User   user.User
	Moment time.Time
}

// SimpleAudit captures the first time a record was written as well
// as the last time the record was updated. The first time a record
// is written First and Last will be identical.
type SimpleAudit struct {
	First Audit `json:"first"`
	Last  Audit `json:"last"`
}

// FromRequest is a convenience function that retrieves the App
// and User structs from the request context. The moment is also
// set to time.Now
func FromRequest(r *http.Request) (Audit, error) {
	var (
		a   app.App
		u   user.User
		err error
	)

	a, err = app.FromRequest(r)
	if err != nil {
		return Audit{}, err
	}

	u, err = user.FromRequest(r)
	if err != nil {
		return Audit{}, err
	}

	return Audit{App: a, User: u, Moment: time.Now()}, nil
}
