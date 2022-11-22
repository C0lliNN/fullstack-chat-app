package marshaler

import "encoding/json"

type JSONMarshaller struct{}

func NewJSONMarshaller() JSONMarshaller {
	return JSONMarshaller{}
}

func (m JSONMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (m JSONMarshaller) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
