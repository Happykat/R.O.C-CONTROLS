package roc

import (
	"github.com/hybridgroup/gobot/api"
)

func (roc *Roc) apiCreate() {

	a := api.NewAPI(roc.gbot)
	a.Debug()
	a.Start()
}

func (roc *Roc) controlBind() {

	//TODO Video Server call
}