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
	// json: cannot unmarshal "bad value" into Go value of type jsawn_test.Custom
	// {"first":"foo","second":"not missing anymore","third":""}
}

type testCase struct {
	Raw  []byte
	Want interface{}
	Got  interface{}
}

type subStruct struct {
	FirstName  string `json:"fname"`
	MiddleName string `json:"mname" jsawn:"optional"`
	LastName   string `json:"lname"`
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

	t.Run("work with string", func(t *testing.T) {
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

	t.Run("work with int", func(t *testing.T) {
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

	t.Run("work with time", func(t *testing.T) {
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

	t.Run("work with structs", func(t *testing.T) {
		wantTime, _ := time.Parse(time.RFC3339, "2022-01-10T16:07:37+01:00")
		data := []struct {
			First  string     `json:"first"`
			Second int        `json:"second"`
			Third  time.Time  `json:"third"`
			Fourth subStruct  `json:"fourth"`
			Fifth  *subStruct `json:"fifth"`
		}{{
			"foo", 42, wantTime,
			subStruct{FirstName: "foo", LastName: "bar"},
			&subStruct{FirstName: "foo", LastName: "bar"},
		}, {
			"", 0, time.Now(),
			subStruct{},
			nil,
		}}

		want := data[0]
		got := data[1]

		jsonStr := []byte(`{
			"first": "foo",
			"second": 42,
			"third": "2022-01-10T16:07:37+01:00",
			"fourth": {
				"fname": "foo",
				"mname": 42,
				"lname": "bar"
			},
			"fifth": {
				"fname": "foo",
				"mname": 42,
				"lname": "bar"
			}
		}`)

		err := jsawn.Unmarshal(jsonStr, &got)
		if err != nil {
			var parseWarn *jsawn.ParseWarning
			if errors.As(err, &parseWarn) {
				fmt.Println(parseWarn)
			} else {
				t.Error(err)
			}
		}

		if want.First != got.First ||
			want.Second != got.Second ||
			want.Third.Format(time.RFC3339) != got.Third.Format(time.RFC3339) ||
			want.Fourth != got.Fourth ||
			!reflect.DeepEqual(want.Fifth, got.Fifth) {

			t.Errorf("\nwant:\n%+v\ngot:\n%+v\n", want, got)
		}
	})
}
