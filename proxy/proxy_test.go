package proxy

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundRobin(t *testing.T) {
	handler := func(resp string) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, resp)
		}
	}

	sp := NewProxy()

	fruits := []string{"apple", "mango", "strawberry"}
	for _, f := range fruits {
		up := httptest.NewServer(http.HandlerFunc(handler(f)))
		sp.AddService(up.URL)
	}

	ts := httptest.NewServer(http.HandlerFunc(sp.handleProxy))
	defer ts.Close()

	assertRoundRobin(t, ts.URL, fruits[0])
	assertRoundRobin(t, ts.URL, fruits[1])
	assertRoundRobin(t, ts.URL, fruits[2])
	assertRoundRobin(t, ts.URL, fruits[0])
}

func assertRoundRobin(t *testing.T, url, expected string) {
	res, err := http.Get(url)
	assert.NoError(t, err)

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, expected, string(greeting))
}

func BenchmarkProxy(b *testing.B) {
	for i := 1; i <= 8; i++ {
		b.Run(fmt.Sprintf("%d upstream", i), func(b *testing.B) {
			runBench(b, i)
		})
	}
}

func runBench(b *testing.B, nupstream int) {
	sp := NewProxy()
	sp.disableStartupLog = true
	go sp.Start()

	for i := 0; i < nupstream; i++ {
		up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(nil)
		}))
		sp.AddService(up.URL)
	}

	for i := 0; i < b.N; i++ {
		res, _ := http.Get("http://localhost:9000")
		res.Body.Close()
	}

	sp.Stop(context.TODO())
}
