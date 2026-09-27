package web

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticFS(t *testing.T) {
	hfs, err := StaticFS()
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.FileServer(hfs))
	defer srv.Close()

	for _, p := range []string{"/", "/index.html", "/style.css", "/main.js", "/polygon.html"} {
		resp, err := http.Get(srv.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("%s -> %d len=%d\n", p, resp.StatusCode, len(body))
	}
}
