package service

import (
	"net/http"

	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/rs/zerolog"
)

type Authorizer interface {
	Authorize(lgr zerolog.Logger, r *http.Request, sub audit.Audit) error
}

type AuthorizeService struct {
	Authorizer Authorizer
}

func (aas AuthorizeService) Authorize(lgr zerolog.Logger, r *http.Request, sub audit.Audit) error {
	return aas.Authorizer.Authorize(lgr, r, sub)
}
