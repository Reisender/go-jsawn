# JSAWN (JAY-sawn)

This is a JSON library to add to the capabilities of the standard 'encoding/json' library.

## Unmarshalling

The first enhancement is to json.Unmarshal. The test file shows an example of the problem
and how this library addresses it.

At a high level, the problem is that basic types on struct fields can fail to parse from JSON
without causing other fields to not parse. If there is a custom struct field that fails to parse,
the *other fields of the struct don't get parsed*. This means that if there are optional struct
fields, any parsing errors on those will cause the entire parse to fail. This utility allows it
to continue to parse everything else and return the failed field(s) as parse warnings.

### `jsawn="optional"`

The jsawn.Unmarshal() returns a specific error type `ParseWarning` if the only parse errors
happened on "optional" fields. If there was an error on a non-optional field, that error
is returned instead of the warnings.

```golang
type Person struct {
  FirstName string `json:"fname"`
  LastName string `json:"lname"`

  // optional field denoted by the jsawn tag
  HairColor color.RGBA `json:"hair_color" jsawn:"optional"`
}
```

When there is a problem parsing this:

```json
{
  "color":"#FFFFFF",
  "fname":"Foo",
  "lname":"Bar"
}
```

... the FirstName and LastName values will parse and there will be a warning for the HairColor failure.
You can inspect the warnings like this:

```golang
var p Person
err := jsawn.Unmarshal(jsonStr, &p)
if err != nil {
  var pw *jsawn.ParseWarning
  if errors.As(err, &pw) {
    for _, warning := range pw.Warnings {
      // warning is the original *json.UnmarshalTypeError
      log.Warn(warning)
    }
  }
}
```

### `jsawn="required"`

If you want to make sure a certain field is present in the JSON, you can use the
"required" tag value. If a required field is not present in the JSON string,
a \*json.UnmarshalTypeError will be returned.
