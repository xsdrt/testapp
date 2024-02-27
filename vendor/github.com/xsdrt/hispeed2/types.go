package hispeed2

import "database/sql"

type initPaths struct {
	rootPath    string
	folderNames []string
}

type cookieConfig struct {
	name     string
	lifetime string
	persist  string
	secure   string
	domain   string
}

type databaseConfig struct {
	dsn      string
	database string
}

type Database struct { //Exporting
	DataType string //I.E. Postgres, MariaDb etc...
	Pool     *sql.DB
}
