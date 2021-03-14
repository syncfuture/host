package siris

import "github.com/kataras/iris/v12"

type (
	Action struct {
		Route      string
		Area       string
		Controller string
		Action     string
		Handler    iris.Handler
	}
)
