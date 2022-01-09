package jsawn_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"go-jsawn"
)

type Custom string

// Always error on unmarshal for this type to show problem
func (c *Custom) UnmarshalJSON(raw []byte) error {
	return &json.UnmarshalTypeError{
		Value: `-> ` + string(raw) + ` <-`,
		Type:  reflect.TypeOf(c),
	}
}

func Example_theProblem() {

	data := struct {
		First  string `json:"first"`
		Second string `json:"second"` // doesn't parse due to Custom parse failure
		Third  Custom `json:"third"`  // Custom type that fails to parse
	}{}

	// third comes before second and errors on parse
	jsonStr := []byte(`{
		"first": "foo",
		"third": "bad value",
		"second": "baz"
	}`)

	err := json.Unmarshal(jsonStr, &data)
	if err != nil {
		fmt.Println(err.Error())
	}

	newStr, err := json.Marshal(&data)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(newStr))

	// Output:
	// json: cannot unmarshal -> "bad value" <- into Go struct field .third of type *jsawn_test.Custom
	// {"first":"foo","second":"","third":""}
}

func ExampleUnmarshal() {
	data := struct {
		First  string `json:"first"`
		Second string `json:"second"`
		Third  Custom `json:"third" jsawn:"optional"` // Custom type that fails to parse
	}{}

	// third comes before second and errors on parse
	jsonStr := []byte(`{
		"first": "foo",
		"third": "bad value",
		"second": "baz"
	}`)

	err := jsawn.Unmarshal(jsonStr, &data)
	var parseWarn *jsawn.ParseWarning
	if errors.As(err, &parseWarn) {
		fmt.Println(parseWarn)
	} else if err != nil {
		fmt.Println(err.Error())
	}

	newStr, err := json.Marshal(&data)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(newStr))

	// Output:
	// parse-warning: json: cannot unmarshal "bad value" into Go value of type jsawn_test.Custom
	// {"first":"foo","second":"baz","third":""}
}
