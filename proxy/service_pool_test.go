package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, client")
	}))

	sp := NewServiceProxy()
	sp.AddServer(upstream.URL)

	ts := httptest.NewServer(http.HandlerFunc(sp.Handler))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(t, err)

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, "Hello, client", string(greeting))
}
