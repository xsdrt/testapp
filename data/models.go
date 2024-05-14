package data

import (
	"database/sql"
	"fmt"
	"os"

	db2 "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

// implement a structuer for models in the Hispeed2 framework that coraspond to tables and data stored in a data base...
var db *sql.DB
var upper db2.Session

// Models is the wrapper for all database models
type Models struct {
	// Any models inserted here (and in the New function)
	// are eaisly accessible throughout the entire application.
	Users  User
	Tokens Token
}

// New initializezes the models package for use...
func New(databasePool *sql.DB) Models {
	db = databasePool

	switch os.Getenv("DATABASE_TYPE") {
	case "mysql", "mariadb":
		upper, _ = mysql.New(databasePool)
	case "postgres", "postgreql":
		upper, _ = postgresql.New(databasePool)
	default:
		// do nothing
	}

	return Models{
		Users:  User{},
		Tokens: Token{},
	}
}

func getInsertID(i db2.ID) int {
	idType := fmt.Sprintf("%T", i)
	if idType == "int64" { // If int64, then postgres; return i cast to a int64...
		return int(i.(int64))
	}

	return i.(int) // Otherwise just return i cast to an int...
}
