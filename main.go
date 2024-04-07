package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/akerl/go-lambda/mux"
)

var (
	c *config

	randomRegex  = regexp.MustCompile(`^/random$`)
	faviconRegex = regexp.MustCompile(`^/favicon.ico$`)
	indexRegex   = regexp.MustCompile(`^/$`)
)

func main() {
	if err := loadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	d := mux.NewDispatcher(
		mux.NewRoute(randomRegex, randomHandler),
		mux.NewRoute(faviconRegex, faviconHandler),
		mux.NewRoute(indexRegex, indexHandler),
	)
	mux.Start(d)
}
