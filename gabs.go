package cloudy

import (
	"encoding/json"

	"github.com/Jeffail/gabs/v2"
)

func ToGabs(item interface{}) (*gabs.Container, error) {
	data, err := json.MarshalIndent(item, "", "   ")
	if err != nil {
		return nil, err
	}
	return gabs.ParseJSON(data)
}

func FromGabs(c *gabs.Container, v interface{}) error {
	return json.Unmarshal(c.Bytes(), v)
}
