package main

import "net/http"

var (
	_ http.Handler = http.HandlerFunc((&controller{}).healthz)
	_ middleware   = (&controller{}).logging
	_ middleware   = (&controller{}).tracing
	_ middleware   = (&controller{}).authorization
)
