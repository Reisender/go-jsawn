package jsawn

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
)

// the jsawn tag values for struct fields
const (
	TagOptional = "optional"
	TagRequired = "required"
)

func Unmarshal(data []byte, val interface{}) error {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(val)}
	}

	v := rv.Elem() // deref the val pointer
	vt := v.Type()

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
		if parts := strings.Split(jsonTag, ","); len(parts) > 0 {
			jsonTag = parts[0] // get the first part in case of json:"name,omitempty"
		}

		var newVal reflect.Value

		// if the field type is a ptr, deref it for the "New()" call
		if f.Type().Kind() == reflect.Ptr {
			newVal = reflect.New(f.Type().Elem()) // New() returns ptr to new val of f.Type()
		} else {
			newVal = reflect.New(f.Type()) // New() returns ptr to new val of f.Type()
		}

		if raw, ok := rawMap[jsonTag]; ok {
			err := Unmarshal(raw, newVal.Interface()) // recursive
			if err != nil {
				var parseWarn *ParseWarning
				if errors.As(err, &parseWarn) {
					// add nested warnings
					for _, warn := range parseWarn.Warnings {
						if tErr, ok := warn.(*json.UnmarshalTypeError); ok {
							tErr.Field = ft.Name + "." + tErr.Field
							tErr.Struct = vt.Name()
						}
						pwarn.Warnings = append(pwarn.Warnings, warn)
					}
				} else {
					// not a nested warning so see if it is err on optional field
					if customJsonTag != TagOptional {
						// return the err since this field is not optional

						// update struct and field since the error was nested
						if tErr, ok := err.(*json.UnmarshalTypeError); ok {
							tErr.Field = ft.Name + "." + tErr.Field
							tErr.Struct = vt.Name()
							return tErr
						}

						return err
					}

					// is warning err so capture and move on
					var parseErr *json.UnmarshalTypeError
					if errors.As(err, &parseErr) {
						parseErr.Field = ft.Name
						parseErr.Struct = vt.Name()
						pwarn.Warnings = append(pwarn.Warnings, parseErr)
					} else {
						// some other err so make our own type error from it
						pwarn.Warnings = append(pwarn.Warnings, &json.UnmarshalTypeError{
							Value:  "value",
							Type:   ft.Type,
							Struct: vt.Name(),
							Field:  ft.Name,
						})
					}
					continue
				}
			}

			if !f.CanSet() {
				panic("can't set field")
			}

			// if the field type was a point, don't deref the newVal
			if f.Type().Kind() == reflect.Ptr {
				f.Set(newVal)
			} else {
				f.Set(newVal.Elem())
			}
		} else if customJsonTag == TagRequired {
			// missing required field so return error
			return &json.UnmarshalTypeError{
				Value:  "missing required field",
				Type:   ft.Type,
				Struct: vt.Name(),
				Field:  ft.Name,
			}
		}
	}

	if len(pwarn.Warnings) > 0 {
		return pwarn
	}

	return nil
}
