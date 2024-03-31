package main

import (
	"testapp/data"
	"testapp/handlers"
	"testapp/middleware"

	"github.com/xsdrt/hispeed2"
)

type application struct {
	App        *hispeed2.HiSpeed2
	Handlers   *handlers.Handlers
	Models     data.Models
	Middleware *middleware.Middleware
}

func main() {
	h := initApplication()
	h.App.ListenAndServe()
}
