package gosocket

import "encoding/json"

type Data []byte

func (d Data) Receive(v interface{}) error {
	return json.Unmarshal([]byte(d), v)
}
