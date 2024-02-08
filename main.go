package main

import (
	"testapp/handlers"

	"github.com/xsdrt/hispeed2"
)

type application struct {
	App      *hispeed2.HiSpeed2
	Handlers *handlers.Handlers
}

func main() {
	h := initApplication()
	h.App.ListenAndServe()
}
