package piwik_test

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"gopkg.in/macaron.v1"
)

type piwikServer struct {
	*httptest.Server
	lastparams url.Values
}

func createPiwikServer() *piwikServer {
	m := macaron.New()
	m.Get("/piwik.php", func(ctx *macaron.Context) {

	})
}

func TestPiwik(t *testing.T) {

}
