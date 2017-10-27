package piwik

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/macaron.v1"
)

// Options configures the piwik middleware
type Options struct {
	// The URL of your piwik installation (with our without /piwik.php)
	PiwikURL string

	// Ignore the Do not Track header that is sent by the browser. This is not recommended
	IgnoreDoNotTrack bool

	// The ID of the website in piwik
	WebsiteID string

	// The piwik API's access token
	Token string
}

func prepareOptions(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}
	opt.PiwikURL = strings.TrimSuffix(strings.TrimSuffix(opt.PiwikURL, "piwik.php"), "/") + "/piwik.php?"

	return opt
}

// Piwik returns a new macaron.Handler that sends every request to the piwik tracking API
func Piwik(options ...Options) macaron.Handler {
	opt := prepareOptions(options)

	return func(ctx *macaron.Context, logger *log.Logger) {
		h := ctx.Req.Header
		if !opt.IgnoreDoNotTrack && h.Get("DNT") == "1" {
			return
		}

		params := make(url.Values)
		params.Set("idsite", opt.WebsiteID)
		params.Set("rec", "1")

		proto := h.Get("X-Forwarded-Proto")
		if proto == "" {
			if ctx.Req.TLS != nil {
				proto = "https"
			} else {
				proto = "http"
			}
		}
		host := h.Get("X-Forwarded-Host")
		if host == "" {
			host = ctx.Req.Host
		}

		params.Set("url", proto+"://"+host+ctx.Req.URL.String())
		params.Set("apiv", "1")
		params.Set("urlref", h.Get("Referer"))
		params.Set("ua", h.Get("User-Agent"))
		params.Set("lang", h.Get("Accept-Language"))

		ip := ctx.RemoteAddr()
		if strings.Contains(ip, ",") {
			ipv6 := strings.Split(ip, ",")
			ip = strings.TrimPrefix(strings.TrimSpace(ipv6[0]), "::ffff:")
		}
		params.Set("token_auth", opt.Token)
		params.Set("cip", ip)

		// collecting data is finished, go async now
		go func() {
			res, err := http.Get(opt.PiwikURL + params.Encode())
			if err != nil {
				logger.Println("Error contacting piwik:", err)
			}
			if res.StatusCode != http.StatusOK {
				logger.Println("Error contacting piwik:", res.Status)
			}
		}()

		ctx.Next()
	}
}
