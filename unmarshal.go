package jsawn

import (
	"encoding/json"
	"reflect"
)

// that jsawn tag value for optional fields
const TagOptional = "optional"

type ParseWarning struct {
	json.UnmarshalTypeError
}

func (w ParseWarning) Error() string {
	return "parse-warning: " + w.UnmarshalTypeError.Error()
}

func Unmarshal(data []byte, val interface{}) error {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(val)}
	}

	rawMap := map[string]json.RawMessage{}

	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}

	v := reflect.ValueOf(val).Elem() // get the elem of the val pointer

	var pwarn *ParseWarning

	// look at each field and parse and
	// don't error out on optional fields
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := v.Type().Field(i)

		customJsonTag := ft.Tag.Get("jsawn")
		jsonTag := ft.Tag.Get("json")

		newVal := reflect.New(f.Type()) // New() returns ptr to new val of f.Type()

		if raw, ok := rawMap[jsonTag]; ok {
			err := json.Unmarshal(raw, newVal.Interface())
			if err != nil {
				if customJsonTag != TagOptional {
					return err
				}

				// is warning err so capture and move on
				pwarn = &ParseWarning{json.UnmarshalTypeError{
					Value: string(raw),
					Type:  ft.Type,
				}}
				continue
			}

			if !f.CanSet() {
				panic("can't set field")
			}
			f.Set(newVal.Elem())
		}
	}

	return pwarn
	//return &ParseWarning{}
}
