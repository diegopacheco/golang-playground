# go-essentials-json

JSON with the standard library `encoding/json`: struct tags, `Marshal`/`Unmarshal`, pretty printing with `MarshalIndent`, `time.Time` encoding, and streaming decode.

### How it works

`src/main.go` marshals a tagged struct, unmarshals JSON back into one, pretty prints a nested map with `MarshalIndent`, encodes a `time.Time` field as RFC 3339, and uses a `json.Decoder` to stream several concatenated JSON objects from one reader.

### Run

```bash
./run.sh
```

### Output

```
marshal: {"name":"alice","age":30,"roles":["admin","user"],"active":true}
unmarshal: {Name:bob Age:25 Roles:[user] Active:false}
prettyPrint:
{
  "a": 1,
  "b": 2,
  "nested": {
    "x": 10,
    "y": 20
  }
}
customTime: {"name":"login","at":"2026-07-12T09:30:00Z"}
stream: [id=1 id=2 id=3]
```
