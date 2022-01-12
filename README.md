# JSAWN (JAY-sawn)

This is a JSON library to add to the capabilities of the standard 'encoding/json' library.

### Unmarshalling

The first enhancement is to json.Unmarshal. The test file shows an example of the problem
and how this library addresses it.

At a high level, the problem is that basic types on struct fields can fail to parse from JSON
without causing other fields to not parse. If there is a custom struct field that fails to parse,
the other fields of the struct don't get parsed. This means that if there are optional struct
fields, any parsing errors should not cause the entire parse to fail. Rather, it should continue
to parse everything else and return the failed field(s) as parse warnings.

The jsawn.Unmarshal() returns a specific error type `ParseWarning` if the only parse errors
happened on "optional" fields. If there was an error on a non-optional field, that error
is returned instead of the warnings.
