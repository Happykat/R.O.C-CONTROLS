package roc

import (
	"errors"
	"log"

	"github.com/Remote-Oculus-Controller/R.O.C-CONTROLS/network"
	"github.com/Remote-Oculus-Controller/proto"
	"github.com/hybridgroup/gobot"
)

//Robot defined all element needed to correctly create a robot compatible with the architecture.
type Robot struct {
	*gobot.Robot
	l    *network.LRocNet
	cmap map[uint32]func(*rocproto.Packet) error
}

//NewRocRobot create a new shell for a "robots" to be include
func NewRocRobot(l *network.LRocNet) *Robot {

	r := new(Robot)
	r.l = l
	r.cmap = make(map[uint32]func(*rocproto.Packet) error)
	return r
}

//Send force header sender section and check if packet can be and was sent.
func (r *Robot) Send(p *rocproto.Packet) error {

	p.Header = p.Header | (uint32(rocproto.Packet_CONTROL_SERVER) << uint32(rocproto.Packet_SHIFT_SEND))
	if r.l == nil {
		e := errors.New("Linker not set, cannot send rocproto.Packet")
		log.Println(e)
		return e
	}
	r.l.Send(p)
	return nil
}

//AddFunc can add a function to the command method slice (r.cmap[]) and/or to the api.
//
//A code is necessary in case of a normal function, used as the command ID.
//Giving a name is not mandatory put highly advised, Mandatory for api entry.
func (r *Robot) AddFunc(f func(*rocproto.Packet) error, code uint32, api func(map[string]interface{}) interface{}, name string) {
	if f != nil && code != 0 {
		log.Println("Assigning function", name, "to code", code)
		_, k := r.cmap[code]
		if k {
			log.Println("Code", code, "already assigned, override")
		}
		r.cmap[code] = f
	}
	if api != nil {
		log.Println("Creating api entry", name)
		r.AddCommand(name, api)
	}
}
