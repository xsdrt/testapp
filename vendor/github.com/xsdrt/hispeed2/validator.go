package hispeed2

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

type Validation struct {
	Data   url.Values        // used for form posts
	Errors map[string]string // place to store errors...
}

func (h *HiSpeed2) Validator(data url.Values) *Validation {
	return &Validation{
		Errors: make(map[string]string),
		Data:   data,
	}
}

func (v *Validation) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validation) AddError(key, message string) { // add errors to the map...
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validation) Has(field string, r *http.Request) bool { // for a form post...
	x := r.Form.Get(field) // check to see if a value(x) is in the form post...
	if x == "" {
		return false
	}
	return true
}

func (v *Validation) Required(r *http.Request, fields ...string) { // check for required fields...
	for _, field := range fields {
		value := r.Form.Get(field)
		if strings.TrimSpace(value) == "" { // if after blank (empty spaces) are trimmed out and still blank then error...
			v.AddError(field, "This field cannot be blank")
		}
	}
}

func (v *Validation) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message) //  key is the name of the field in question...
	}
}

func (v *Validation) IsEmail(field, value string) {
	if !govalidator.IsEmail(value) {
		v.AddError(field, "Requires a valid email address")
	}
}

func (v *Validation) IsInt(field, value string) {
	_, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, "This field must be an integer")
	}
}

func (v *Validation) IsFloat(field, value string) {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		v.AddError(field, "This field must be a floating point number")
	}
}

func (v *Validation) IsDateISO(field, value string) {
	_, err := time.Parse("2024-04-15", value)
	if err != nil {
		v.AddError(field, "This field must be a date in the ISO form YYYY-MM-DD")
	}
}

func (v *Validation) NoSpaces(field, value string) {
	if govalidator.HasWhitespace(value) {
		v.AddError(field, "White/Blank spaces are not permitted")
	}
}
