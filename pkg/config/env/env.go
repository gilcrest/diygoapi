package env

import "database/sql"

type Env struct {
	Db *sql.DB
}
