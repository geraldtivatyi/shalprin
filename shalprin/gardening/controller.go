package gardening

import (
	"net/http"
)

func (d *Device) List() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		v := NewView(w)

		m.Data(NewList())
		v.JSON(m)
	})
}

func (d Device) Update() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := NewModel(r, d.Store)
		v := NewView(w)

		v.JSON(m)
	})
}
