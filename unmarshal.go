package jsawn

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

// that jsawn tag value for optional fields
const TagOptional = "optional"

type ParseWarning struct {
	Warnings []*json.UnmarshalTypeError
}

func (w ParseWarning) Error() string {
	// pick the first one for now
	warnings := ""
	newline := ""
	for _, warn := range w.Warnings {
		warnings = fmt.Sprintf("%s%s%s", warnings, newline, warn.Error())
		newline = "\n"
	}

	plurality := "warning"
	if len(w.Warnings) > 1 {
		plurality += "s"
	}

	return fmt.Sprintf("%d parse %s\n%s",
		len(w.Warnings),
		plurality,
		warnings,
	)
}

func Unmarshal(data []byte, val interface{}) error {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(val)}
	}

	v := rv.Elem() // get the elem of the val pointer
	//vi := reflect.Indirect(rv.Elem())
	vt := v.Type()

	if v.Kind() == reflect.Ptr {
		//v = vi
	}

	// hand off to standard unmarshal if not a struct
	// or a struct pointer
	if v.Kind() != reflect.Struct {
		// not a struct so process like normal
		return json.Unmarshal(data, val)
	}

	// use custom Unmarshal if it exists
	if _, ok := val.(json.Unmarshaler); ok {
		return json.Unmarshal(data, val)
	}

	rawMap := map[string]json.RawMessage{}

	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}

	// collect any parse warnings
	// which are errors that happen on optional fields
	pwarn := &ParseWarning{}

	// look at each field and parse and
	// don't error out on optional fields
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := vt.Field(i)

		customJsonTag := ft.Tag.Get("jsawn")
		jsonTag := ft.Tag.Get("json")

		newVal := reflect.New(f.Type()) // New() returns ptr to new val of f.Type()

		if jsonTag == "fifth" {
			runtime.Breakpoint()
		}

		if raw, ok := rawMap[jsonTag]; ok {
			err := Unmarshal(raw, newVal.Interface()) // recursive
			if err != nil {
				var parseWarn *ParseWarning
				if errors.As(err, &parseWarn) {
					// add nested warnings
					for _, warn := range parseWarn.Warnings {
						warn.Field = ft.Name + "." + warn.Field
						warn.Struct = vt.Name()
						pwarn.Warnings = append(pwarn.Warnings, warn)
					}
				} else {
					// not a nested warning so see if it is a warning or err
					if customJsonTag != TagOptional {
						return err
					}

					// is warning err so capture and move on
					var parseErr *json.UnmarshalTypeError
					if errors.As(err, &parseErr) {
						parseErr.Field = jsonTag
						parseErr.Struct = vt.Name()
						pwarn.Warnings = append(pwarn.Warnings, parseErr)
					} else {
						pwarn.Warnings = append(pwarn.Warnings, &json.UnmarshalTypeError{
							Value:  string(raw),
							Type:   ft.Type,
							Struct: vt.Name(),
							Field:  jsonTag,
						})
					}
					continue
				}
			}

			if !f.CanSet() {
				panic("can't set field")
			}
			f.Set(newVal.Elem())
		}
	}

	if len(pwarn.Warnings) > 0 {
		return pwarn
	}

	return nil
}
