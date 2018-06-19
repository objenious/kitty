package kitty_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/objenious/kitty"
	"github.com/objenious/kitty/gorilla"
)

func ExampleServer() {
	type fooRequest struct{ Name string }
	type fooResponse struct{ Message string }

	foo := func(ctx context.Context, request interface{}) (interface{}, error) {
		fr := request.(fooRequest)
		return fooResponse{Message: fmt.Sprintf("Good morning %s !", fr.Name)}, nil
	}

	decodeFooRequest := func(ctx context.Context, r *http.Request) (interface{}, error) {
		var request fooRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return nil, err
		}
		return request, nil
	}

	kitty.NewServer().Config(kitty.Config{HTTPPort: 8081}).
		Router(gorilla.Router()).
		HTTPEndpoint(foo, kitty.Method("POST"), kitty.Path("/foo"), kitty.Decoder(decodeFooRequest)).
		Run(context.Background())
}
