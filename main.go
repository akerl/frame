package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/akerl/go-lambda/mux"
)

var (
	c *config

	randomRegex = regexp.MustCompile(`^/random$`)
	indexRegex  = regexp.MustCompile(`^/$`)
)

func main() {
	if err := loadConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	d := mux.NewDispatcher(
		mux.NewRouteWithBasicAuth(randomRegex, randomHandler, c.Users),
		mux.NewRouteWithBasicAuth(indexRegex, indexHandler, c.Users),
	)
	mux.Start(d)
}
