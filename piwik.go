package piwik

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/macaron.v1"
)

type Options struct {
	PiwikUrl         string
	IgnoreDoNotTrack bool
	WebsiteID        string
	Token            string
}

func prepareOptions(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}
	opt.PiwikUrl = strings.TrimSuffix(strings.TrimSuffix(opt.PiwikUrl, "piwik.php"), "/") + "/piwik.php?"

	return opt
}

func Piwik(options ...Options) macaron.Handler {
	opt := prepareOptions(options)

	return func(ctx *macaron.Context, logger *log.Logger) {
		params := make(url.Values)
		params.Set("idsite", opt.WebsiteID)
		params.Set("rec", "1")

		params.Set("url", ctx.Req.URL.String())
		params.Set("apiv", "1")

		h := ctx.Req.Header
		params.Set("urlref", h.Get("Referer"))
		fmt.Println("USERAGENT")
		fmt.Println(h.Get("Accept-Language"))
		params.Set("ua", h.Get("User-Agent"))
		params.Set("lang", h.Get("Accept-Language"))

		params.Set("token_auth", opt.Token)
		params.Set("cip", ctx.RemoteAddr())

		// collecting data is finished, go async now
		go func() {
			logger.Println(params.Encode())
			res, err := http.Get(opt.PiwikUrl + params.Encode())
			if err != nil {
				logger.Println("Error contacting piwik:", err)
			}
			r, err := ioutil.ReadAll(res.Body)
			logger.Println(string(r))
		}()

		ctx.Next()
	}
}
