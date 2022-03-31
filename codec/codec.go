package codec

import (
	"encoding/json"
)

// DefaultCodec is the default codec used by nrpc.
var DefaultCodec Codec = &JSONCodec{}

// Codec is the interface that wraps the nrpc Message data encoding method.
type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// Wraps std json.
type JSONCodec struct{}

// Wraps std json.Marshal.
func (j *JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Wraps std json.Unmarshal
func (j *JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Sets default codec instance.
func SetCodec(c Codec) {
	DefaultCodec = c
}
