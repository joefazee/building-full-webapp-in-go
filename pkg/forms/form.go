package forms

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Form struct {
	url.Values
	Errors errors
}

type customValidator func(field string, value string, obj *Form)

func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

func (f *Form) Required(fields ...string) *Form {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
	return f
}

func (f *Form) MaxLength(field string, d int) *Form {
	value := f.Get(field)
	if value == "" {
		return f
	}

	if utf8.RuneCountInString(value) > d {
		f.Errors.Add(field, fmt.Sprintf("This field is too long (maximum is %d characters)", d))
	}

	return f
}

func (f *Form) PermittedValues(field string, opts ...string) *Form {
	value := f.Get(field)
	if value == "" {
		return f
	}

	for _, opt := range opts {
		if value == opt {
			return f
		}
	}

	f.Errors.Add(field, "This field is not valid")
	return f
}

func (f *Form) Custom(field string, validator customValidator) *Form {
	validator(field, f.Get(field), f)
	return f
}

func (f *Form) Fail(field, msg string) {
	f.Errors.Add(field, msg)
}

func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Implement a MinLength method to check that a specific field in the form
// contains a minimum number of characters. If the check fails then add the
// appropriate message to the form errors.
func (f *Form) MinLength(field string, d int) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if utf8.RuneCountInString(value) < d {
		f.Errors.Add(field, fmt.Sprintf("This field is too short (minimum is %d characters)", d))
	}
}

// Implement a MatchesPattern method to check that a specific field in the form
// matches a regular expression. If the check fails then add the
// appropriate message to the form errors.
func (f *Form) MatchesPattern(field string, pattern *regexp.Regexp) {
	value := f.Get(field)
	if value == "" {
		return
	}
	if !pattern.MatchString(value) {
		f.Errors.Add(field, "This field is invalid")
	}
}
