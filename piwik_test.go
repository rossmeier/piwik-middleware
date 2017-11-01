package piwik

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/macaron.v1"
)

type piwikServer struct {
	*httptest.Server
	lastparams url.Values
}

func (p *piwikServer) p(key string) string {
	return p.lastparams.Get(key)
}

const token = "56ced3748e5df1b6be1e5c56aca45e7a"
const ua = "thisisauseragent"

func createPiwikServer() *piwikServer {
	p := new(piwikServer)
	m := macaron.New()
	m.Get("/piwik.php", func(ctx *macaron.Context) {
		fmt.Println(ctx.Req.URL.Query())
		p.lastparams = ctx.Req.URL.Query()
	})
	p.Server = httptest.NewServer(m)
	return p
}

func check(p *piwikServer, url string, ip string) {
	So(p.p("token_auth"), ShouldEqual, token)
	So(p.p("cip"), ShouldEqual, ip)
}

func TestPiwik(t *testing.T) {
	p := createPiwikServer()
	defer p.Close()

	m := macaron.New()
	m.Use(Piwik(Options{
		HTTPClient: p.Client(),
		Token:      token,
		PiwikURL:   "http://example.com/piwik.php",
		WebsiteID:  "1",
	}))
	m.Get("/", func() string {
		return ""
	})
	m.Get("/ignore", Ignore)
	m.Get("/search", func(ctx *macaron.Context, t TrackingParams) {
		t.Search(ctx.Req.Form.Get("q"), "kat", 3)
	})

	Convey("", t, func() {
		m.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "http://example.com/", nil))
		check(p, "http://example.com/", httptest.DefaultRemoteAddr)
	})
}
