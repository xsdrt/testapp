package hispeed2

import (
	"log"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func (h *HiSpeed2) MigrateUp(dsn string) error {
	// windows only://rootPath := filepath.ToSlash(h.RootPath) // Added this due to a problem in the go-migrate pkg

	// windows only:// m, err := migrate.New("file://"+rootPath+"/migrations", dsn) // create /open up the migration file...changed from h.RootPath to correct an interpretation problem...
	m, err := migrate.New("file://"+h.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		log.Println("Error running migration: ", err)
		return err
	}
	return nil
}

func (h *HiSpeed2) MigrateDownAll(dsn string) error {
	// windows only://rootPath := filepath.ToSlash(h.RootPath)

	// windows only://m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	m, err := migrate.New("file://"+h.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil {
		return err
	}

	return nil
}

func (h *HiSpeed2) Steps(n int, dsn string) error {
	// windows only://rootPath := filepath.ToSlash(h.RootPath)

	// windows only://m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	m, err := migrate.New("file://"+h.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(n); err != nil {
		return err
	}

	return nil
}

func (h *HiSpeed2) MigrateForce(dsn string) error { //if you have an error in the migration file, might be marked dirty in the DB , so force the migration...
	// windows only://rootPath := filepath.ToSlash(h.RootPath)

	// windows only://m, err := migrate.New("file://"+rootPath+"/migrations", dsn)
	m, err := migrate.New("file://"+h.RootPath+"/migrations", dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Force(-1); err != nil { // So we will force the migration down 1... allows oportunity to fix  and retry the migration after we fix the problem in our migration file
		return err
	}

	return nil
}
