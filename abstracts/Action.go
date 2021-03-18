package abstracts

import (
	log "github.com/syncfuture/go/slog"
	"github.com/syncfuture/host/shttp"
)

type ActionGroup struct {
	PreHandlers   []shttp.RequestHandler
	Actions       []*Action
	AfterHandlers []shttp.RequestHandler
}

type Action struct {
	Route      string
	Area       string
	Controller string
	Action     string
	Handlers   []shttp.RequestHandler
}

func NewActionGroup(preHandlers []shttp.RequestHandler, actions []*Action, afterHandlers ...shttp.RequestHandler) *ActionGroup {
	return &ActionGroup{
		PreHandlers:   preHandlers,
		Actions:       actions,
		AfterHandlers: afterHandlers,
	}
}

func NewAction(route, area, controller string, handlers ...shttp.RequestHandler) *Action {
	if len(handlers) == 0 {
		log.Fatal("handlers are missing")
	}

	return &Action{
		Route:      route,
		Area:       area,
		Controller: controller,
		Handlers:   handlers,
	}
}

func (x *Action) AppendHandler(handlers ...shttp.RequestHandler) {
	x.Handlers = append(x.Handlers, handlers...)
}
