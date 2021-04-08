package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	up1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, client 1")
	}))

	up2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, client 2")
	}))

	up3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, client 3")
	}))

	sp := NewProxy()
	sp.AddService(up1.URL)
	sp.AddService(up2.URL)
	sp.AddService(up3.URL)

	ts := httptest.NewServer(http.HandlerFunc(sp.handleProxy))
	defer ts.Close()

	for i := 1; i <= 3; i++ {
		res, err := http.Get(ts.URL)
		assert.NoError(t, err)

		greeting, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		assert.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("Hello, client %d", i), string(greeting))
	}

	res, err := http.Get(ts.URL)
	assert.NoError(t, err)

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, "Hello, client 1", string(greeting))
}

func Benchmark4Upstream(b *testing.B) {
	b.Run("200 microsecond/response", func(b *testing.B) {
		sp := NewProxy()
		for i := 0; i < 4; i++ {
			up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Microsecond * 200)
				msg := fmt.Sprintf("Hello, client %d", i)
				fmt.Fprint(w, msg)
			}))

			sp.AddService(up.URL)
		}

		ts := httptest.NewServer(http.HandlerFunc(sp.handleProxy))
		defer ts.Close()

		b.Run("1000 req", func(b *testing.B) {
			for i := 0; i < 1000; i++ {
				res, _ := http.Get(ts.URL)
				res.Body.Close()
			}
		})

		b.Run("10000 req", func(b *testing.B) {
			for i := 0; i < 10000; i++ {
				res, _ := http.Get(ts.URL)
				res.Body.Close()
			}
		})
	})

	b.Run("20 microsecond/response", func(b *testing.B) {
		sp := NewProxy()
		for i := 0; i < 4; i++ {
			up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Microsecond * 200)
				msg := fmt.Sprintf("Hello, client %d", i)
				fmt.Fprint(w, msg)
			}))

			sp.AddService(up.URL)
		}

		ts := httptest.NewServer(http.HandlerFunc(sp.handleProxy))
		defer ts.Close()

		b.Run("1000 req", func(b *testing.B) {
			for i := 0; i < 1000; i++ {
				res, _ := http.Get(ts.URL)
				res.Body.Close()
			}
		})

		b.Run("10000 req", func(b *testing.B) {
			for i := 0; i < 10000; i++ {
				res, _ := http.Get(ts.URL)
				res.Body.Close()
			}
		})
	})
}
