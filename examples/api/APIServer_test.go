package main

import (
	"testing"

	"github.com/syncfuture/go/config"
	"github.com/syncfuture/host"
)

func TestCreateAPIServer(t *testing.T) {
	cp := config.NewJsonConfigProvider()
	var options *host.APIServerOptions
	cp.GetStruct("APIServer", &options)
	host.NewAPIServer(cp, options)
}
