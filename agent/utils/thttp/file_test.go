package thttp

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"testing"
)

func TestUploadFileJSON(t *testing.T) {
	var wg sync.WaitGroup

	stop := make(chan struct{}, 1)

	wg.Add(1)

	addr := "127.0.0.1:50911"
	url := "/test"

	go startServer(&wg, addr, url, h, stop)

	err := UploadFilesJSON("http://"+addr+url, http.MethodPost, []string{"files"}, []string{
		"file.go",
		"util.go",
	},
		map[string][]string{
			"a": {"b", "fjwel"},
			"1": {"2", "3", "4"},
		}, 0, nil)

	fmt.Printf("error: %v\n", err)

	stop <- struct{}{}

	wg.Wait()
}

func processFile(f *multipart.FileHeader) error {
	file, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	fmt.Printf("file name: \"%v\", content: \"%v\"\n", f.Filename, string(b))
	return nil
}

func h(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1024)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("ParseMultipartForm error: %v", err)))
		return
	}
	for k, files := range r.MultipartForm.File {
		fmt.Printf("field: %v\n", k)
		for _, f := range files {
			err = processFile(f)
			if err != nil {
				_, _ = w.Write([]byte(fmt.Sprintf("processFile error: %v", err)))
				return
			}
		}
	}
	for k, v := range r.MultipartForm.Value {
		fmt.Printf("field: %v, values: %v\n", k, v)
	}

	_, _ = w.Write([]byte(`{"code": 0}`))
}

func startServer(wg *sync.WaitGroup, addr, url string, handlerFunc http.HandlerFunc, stop chan struct{}) {
	if wg == nil {
		return
	}
	defer wg.Done()

	handler := http.NewServeMux()
	handler.HandleFunc(url, handlerFunc)
	s := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		fmt.Printf("start listening....\n")
		if err := s.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	select {
	case <-stop:
		_ = s.Shutdown(context.Background())
		return
	}
}
