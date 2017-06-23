// "Environment" package and type to store common environment
// related items - sql db, logger, etc.
package env

import (
	"database/sql"

	"go.uber.org/zap"
)

type Env struct {
	Db     *sql.DB
	Logger *zap.Logger
}
