package thttp

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func TestRequestJSON(t *testing.T) {
	var X interface{}

	addr := "127.0.0.1:50911"
	url := "/test"

	var wg sync.WaitGroup

	stop := make(chan struct{}, 1)

	wg.Add(1)
	go startServer(&wg, addr, url, get, stop)

	err := RequestJSON("http://"+addr+url, http.MethodGet, nil, 0, &X)
	fmt.Printf("%+v\n%v\n", X, err)
}

func get(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`
{
  "code": 0,
  "data": [1,2,3]
}
`))
}
