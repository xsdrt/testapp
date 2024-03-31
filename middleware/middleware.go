package middleware

import (
	"testapp/data"

	"github.com/xsdrt/hispeed2"
)

type Middleware struct {
	App    *hispeed2.HiSpeed2
	Models data.Models
}
