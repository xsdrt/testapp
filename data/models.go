package data

import (
	"database/sql"
	"os"

	db2 "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

// implement a structuer for models in the Hispeed2 framework that coraspond to tables and data stored in a data base...
var db *sql.DB
var upper db2.Session

type Models struct {
	// Any models inserted here (and in the New function)
	// are eaisly accessible throughout the entire application.
}

func New(databasePool *sql.DB) Models {
	db = databasePool

	if os.Getenv("DATABASE_TYPE") == "mysql" || os.Getenv("DATABASE_TYPE") == "mariadb" {
		upper, _ = mysql.New(databasePool)
	} else {
		upper, _ = postgresql.New(databasePool)
	}

	return Models{}
}
