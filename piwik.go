package piwik

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
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

	// htttp client to be used for requests to piwik
	HTTPClient *http.Client

	// The piwik API's access token
	Token string
}

// TrackingParams will be injected into the context and can be used to define
// piwik actions, like an internal search
type TrackingParams struct {
	// custom variables that will be assigned to the action (cvar)
	ActionCVar map[string]string

	// custom variables that will be assigned to the visitor (_cvar)
	VisitorCVar map[string]string

	// when this is set to true, no information will be sent to piwik
	Ignore bool

	params url.Values
}

// Search notifies piwik about a search done. Params can be:
// - the number of search results (int)
// - the catecory of the search (string)
func (t *TrackingParams) Search(keyword string, params ...interface{}) {
	t.params.Set("search", keyword)
	for _, x := range params {
		switch x.(type) {
		case int:
			t.params.Set("search_count", strconv.Itoa(x.(int)))
		case string:
			t.params.Set("search_cat", x.(string))
		}
	}
}

// Ignore provides a handler that will hinder all requests from being tracked
func Ignore(tracker *TrackingParams) {
	tracker.Ignore = true
}

func prepareOptions(options []Options) Options {
	var opt Options
	if len(options) > 0 {
		opt = options[0]
	}
	opt.PiwikURL = strings.TrimSuffix(strings.TrimSuffix(opt.PiwikURL, "piwik.php"), "/") + "/piwik.php?"
	if opt.HTTPClient == nil {
		opt.HTTPClient = http.DefaultClient
	}

	return opt
}

// Piwik returns a new macaron.Handler that sends every request to the piwik tracking API
func Piwik(options ...Options) macaron.Handler {
	opt := prepareOptions(options)

	return func(ctx *macaron.Context, logger *log.Logger) {
		h := ctx.Req.Header
		if !opt.IgnoreDoNotTrack && h.Get("DNT") == "1" {
			ctx.Map(&TrackingParams{
				ActionCVar:  make(map[string]string),
				VisitorCVar: make(map[string]string),
				params:      make(url.Values),
				Ignore:      true,
			})
			ctx.Next()
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

		p := &TrackingParams{
			ActionCVar:  make(map[string]string),
			VisitorCVar: make(map[string]string),
			params:      params,
		}
		ctx.Map(p)
		ctx.Next()

		if p.Ignore {
			return
		}

		if (len(p.ActionCVar)) > 0 {
			b, err := json.Marshal(p.ActionCVar)
			if err != nil {
				logger.Println("Error marshalling ActionCVar:", err)
			} else {
				params.Set("cvar", string(b))
			}
		}
		if (len(p.VisitorCVar)) > 0 {
			b, err := json.Marshal(p.VisitorCVar)
			if err != nil {
				logger.Println("Error marshalling VisitorCVar:", err)
			} else {
				params.Set("_cvar", string(b))
			}
		}

		// collecting data is finished, go async now
		go func() {
			res, err := opt.HTTPClient.Get(opt.PiwikURL + params.Encode())
			if err != nil {
				logger.Println("Error contacting piwik:", err)
			}
			if res.StatusCode != http.StatusOK {
				logger.Println("Error contacting piwik:", res.Status)
			}
		}()
	}
}
