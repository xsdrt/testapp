package main

import (
	"log"
	"os"

	"github.com/xsdrt/hispeed2"
)

func initApplication() *application {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err) //if can't find the wd then just die as something went wrong...
	}

	// init hispeed2
	his := &hispeed2.HiSpeed2{}
	err = his.New(path)
	if err != nil {
		log.Fatal(err) // again something went wrong , so just die...
	}

	his.AppName = "testapp"
	his.Debug = true

	app := &application{
		App: his,
	}

	return app

}
