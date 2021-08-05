package quote

import (
	"encoding/json"
	"errors"
	"net/http"
)

const quotesUrl = "https://api.quotable.io/random"

type Quote struct {
	Content string `json:"content"`
	Author  string `json:"author"`
}

func Random() (Quote, error) {
	r, err := http.Get(quotesUrl)
	if err != nil {
		return Quote{}, errors.New("sorry, quotes are not reachable at this moment :(")
	}
	defer r.Body.Close()

	quote := &Quote{}
	e := json.NewDecoder(r.Body).Decode(quote)
	if e != nil {
		// here we can just retry several times, maybe it's an occasional problem
		return Quote{}, errors.New("sorry, quotes are not reachable at this moment :(")
	}
	return *quote, nil
}
