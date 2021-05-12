package profile

import "github.com/oligoden/chassis/device/model/data"

type results struct {
	Results []string `json:"results"`
	data.Default
}

func NewSearch() *results {
	r := &results{}
	return r
}
