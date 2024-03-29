package jsawn_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"go-jsawn"
)

type Custom string

// Always error on unmarshal for this type to show problem
func (c *Custom) UnmarshalJSON(raw []byte) error {
	return &json.UnmarshalTypeError{
		Value: string(raw),
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
		"second": "missing"
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
	// json: cannot unmarshal "bad value" into Go struct field .third of type *jsawn_test.Custom
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
		"second": "not missing anymore"
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
	// 1 parse warning
	// json: cannot unmarshal "bad value" into Go struct field .Third of type *jsawn_test.Custom
	// {"first":"foo","second":"not missing anymore","third":""}
}

type dataStruct struct {
	Identity
	*Common
	First   string     `json:"first" jsawn:"required"`
	Second  int        `json:"second"`
	Third   time.Time  `json:"third"`
	Fourth  subStruct  `json:"fourth"`
	Fifth   *subStruct `json:"fifth"`
	Sixth   *float32   `json:"sixth"`
	Seventh *int       `json:"seventh" jsawn:"optional"`
}

type subStruct struct {
	FirstName  string   `json:"fname" jsawn:"required"`
	MiddleName string   `json:"mname" jsawn:"optional"`
	LastName   string   `json:"lname"`
	Aliases    []string `json:"aka" jsawn:"optional"`
}

type Identity struct {
	ID int `json:"id"`
}

// Common stuct will be embedded into the data struct
type Common struct {
	Name string `json:"name"`
}

func TestUnmarshal(t *testing.T) {

	t.Run("expect a pointer", func(t *testing.T) {
		var data int
		err := jsawn.Unmarshal([]byte("42"), data) // not a pointer
		_, ok := err.(*json.InvalidUnmarshalError) // should be err
		if !ok {
			t.Errorf("\nwant err type:\n%T\ngot:\n%T\n",
				(*json.InvalidUnmarshalError)(nil),
				err,
			)
		}
	})

	t.Run("with string", func(t *testing.T) {
		var data string
		want := "foo bar"

		err := jsawn.Unmarshal([]byte(`"`+want+`"`), &data)
		if err != nil {
			t.Error(err)
		}

		if data != want {
			t.Errorf("\nwant:\n%s\ngot:\n%s\n", want, data)
		}
	})

	t.Run("with int", func(t *testing.T) {
		var data int
		want := 42

		err := jsawn.Unmarshal([]byte(`42`), &data)
		if err != nil {
			t.Error(err)
		}

		if data != want {
			t.Errorf("\nwant:\n%d\ngot:\n%d\n", want, data)
		}
	})

	t.Run("with omitempty", func(t *testing.T) {
		type data struct {
			Val int `json:"val,omitempty"`
		}
		got := data{}
		want := data{Val: 42}

		err := jsawn.Unmarshal([]byte(`{"val": 42}`), &got)
		if err != nil {
			t.Error(err)
		}

		if got != want {
			t.Errorf("\nwant:\n%d\ngot:\n%d\n", want, got)
		}
	})

	t.Run("with a non-base type", func(t *testing.T) {
		var data time.Time
		want := time.Now()

		err := jsawn.Unmarshal([]byte(`"`+want.Format(time.RFC3339)+`"`), &data)
		if err != nil {
			t.Error(err)
		}

		if data.Format(time.RFC3339) != want.Format(time.RFC3339) {
			t.Errorf("\nwant:\n%v\ngot:\n%v\n", want, data)
		}
	})

	t.Run("with optional struct fields", func(t *testing.T) {
		sixth := float32(43.33)
		wantTime, _ := time.Parse(time.RFC3339, "2022-01-10T16:07:37+01:00")

		want := dataStruct{
			Identity{1},
			&Common{"Alice"},
			"foo", 42, wantTime,
			subStruct{FirstName: "foo", LastName: "bar"},
			&subStruct{FirstName: "foo", LastName: "bar", Aliases: []string{"joe"}},
			&sixth, nil,
		}

		// create a json string with 3 problems on optional fields
		// at multiple levels of nesting
		jsonStr := []byte(`{
			"id": 1,
			"name": "Alice",
			"first": "foo",
			"second": 42,
			"third": "2022-01-10T16:07:37+01:00",
			"fourth": {
				"fname": "foo",
				"lname": "bar",
				"aka": "['joe']"
			},
			"fifth": {
				"fname": "foo",
				"mname": 42,
				"lname": "bar",
				"aka": ["joe"]
			},
			"sixth": 43.33,
			"seventh": "43"
		}`)

		// do the actual parsing
		var got dataStruct
		err := jsawn.Unmarshal(jsonStr, &got)

		if err == nil {
			t.Error("expected parse warnings and got nil for err")
			return
		}

		// check the warnings
		var parseWarn *jsawn.ParseWarning
		if errors.As(err, &parseWarn) {
			if len(parseWarn.Warnings) != 3 {
				t.Errorf("parse warnings:\nwant:\n%d\ngot:\n%d\n%+v\n", 3, len(parseWarn.Warnings), parseWarn.Warnings)
				return
			}

			if parseWarn.Warnings[0].(*json.UnmarshalTypeError).Field != "Fourth.Aliases" {
				t.Errorf("want %s got %s", "Fourth.Aliases", parseWarn.Warnings[0].(*json.UnmarshalTypeError).Field)
			}

			if parseWarn.Warnings[1].(*json.UnmarshalTypeError).Field != "Fifth.MiddleName" {
				t.Errorf("want %s got %s", "Fifth.MiddleName", parseWarn.Warnings[1].(*json.UnmarshalTypeError).Field)
			}

			if parseWarn.Warnings[2].(*json.UnmarshalTypeError).Field != "Seventh" {
				t.Errorf("want %s got %s", "Seventh", parseWarn.Warnings[2].(*json.UnmarshalTypeError).Field)
			}
		} else {
			t.Error(err)
		}

		// expect want and got to be the same
		if !reflect.DeepEqual(want, got) {
			t.Errorf("\nwant:\n%+v\ngot:\n%+v\n", want, got)
		}
	})

	t.Run("with required struct fields", func(t *testing.T) {
		got := dataStruct{}

		// missing the first field which is required
		jsonStr := []byte(`{
			"second": 42
		}`)

		err := jsawn.Unmarshal(jsonStr, &got)

		if err == nil {
			t.Error("expected err for missing required field")
			return
		}

		var parseWarn *jsawn.ParseWarning
		if errors.As(err, &parseWarn) {
			t.Errorf("expected err for missing required field but got warning:\n%s", parseWarn)
			return
		}

		// check the specific error
		var parseErr *json.UnmarshalTypeError
		if errors.As(err, &parseErr) {
			if parseErr.Field != "First" {
				t.Errorf("want %s got %s", "First", parseErr.Field)
			}
		} else {
			t.Error(err)
		}
	})

	t.Run("with required nested struct fields", func(t *testing.T) {
		got := dataStruct{}

		// missing the first field which is required
		jsonStr := []byte(`{
			"first": "foo",
			"fourth": {}
		}`)

		err := jsawn.Unmarshal(jsonStr, &got)

		if err == nil {
			t.Error("expected err for missing required field")
			return
		}

		var parseWarn *jsawn.ParseWarning
		if errors.As(err, &parseWarn) {
			t.Errorf("expected err for missing required field but got warning:\n%s", parseWarn)
			return
		}

		// check the specific error
		var parseErr *json.UnmarshalTypeError
		if errors.As(err, &parseErr) {
			if parseErr.Field != "Fourth.FirstName" {
				t.Errorf("want %s got %s", "Fourth.FirstName", parseErr.Field)
			}
		} else {
			t.Error(err)
		}
	})
}
