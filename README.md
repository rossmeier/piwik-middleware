# piwik-middleware
[![Build Status](https://travis-ci.org/veecue/piwik-middleware.svg?branch=master)](https://travis-ci.org/veecue/piwik-middleware)
[![GoDoc](https://godoc.org/github.com/veecue/piwik-middleware?status.svg)](https://godoc.org/github.com/veecue/piwik-middleware)
[![Go Report Card](https://goreportcard.com/badge/github.com/veecue/piwik-middleware)](https://goreportcard.com/report/github.com/veecue/piwik-middleware)

Piwik is a middleware for [macaron](https://go-macaron.com) that tracks visits via [piwik](https://piwik.org)'s [HTTP Tracking API](https://developer.piwik.org/api-reference/tracking-api)
on the serverside without any visibility for the client.

## Installation
`go get -u github.com/veecue/piwik-middleware`

## Usage
```go
package main

import (
	"github.com/veecue/piwik-middleware"
	"gopkg.in/macaron.v1"
)

func main() {
	m := macaron.Classic()
	m.Use(piwik.Piwik(piwik.Options{
		PiwikURL:  "http://localhost/piwik",
		Token:     "56ced3748e5df1b6be1e5c56aca45e7a",
		WebsiteID: "1",
	}))
	m.Get("/", func() string {
		return "Hi"
	})
	m.Run()
}
```

## Options
The middleware takes a variety of options:
```go
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
```

## Contributing
If you have found any bugs or want to propose a new feature, feel free to submit an issue or PR
