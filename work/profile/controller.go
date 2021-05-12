package profile

import context "context"

func (d Device) Profile(ctx context.Context, in *ProfileMessage) (*ProfileMessage, error) {
	c := d.Store.Connect(uint(in.UserID), []uint{})

	e := NewRecord()
	c.Read(e)
	if c.Err() != nil {
		// m.Err(c.Err)
		// return
	}

	return &ProfileMessage{Address: e.Address}, nil
}

// func (d Device) PredictAddress() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		m := NewModel(r, d.Store)
// 		e := NewSearch()
// 		m.Data(e)
// 		v := NewView(w)

// 		qs, ok := r.URL.Query()["query"]
// 		if !ok || len(qs) == 0 {
// 			m.Err("bad request, require query parameter")
// 			return
// 		}

// 		client, err := maps.NewClient(maps.WithAPIKey("AIzaSyDX69tbU28HetTpfPS0jjj5KhalpGEA6Vc"))
// 		if err != nil {
// 			log.Print(err)
// 			m.Err(err)
// 			return
// 		}

// 		req := &maps.PlaceAutocompleteRequest{
// 			Input:      qs[0],
// 			Components: map[maps.Component][]string{maps.ComponentCountry: []string{"za"}},
// 		}
// 		rsp, err := client.PlaceAutocomplete(context.Background(), req)
// 		if err != nil {
// 			log.Print(err)
// 			m.Err(err)
// 			return
// 		}

// 		for _, p := range rsp.Predictions {
// 			fmt.Println(p.Description)
// 			e.Results = append(e.Results, p.Description)
// 		}

// 		v.JSON(m)
// 	})
// }
