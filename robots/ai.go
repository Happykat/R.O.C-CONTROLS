package robots

import (
	"github.com/Happykat/R.O.C-CONTROLS"
	"github.com/Happykat/R.O.C-CONTROLS/misc"
	"github.com/Happykat/R.O.C-CONTROLS/rocproto"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/gpio"
)

const (
	LOCK = uint32(MAI_TAG)
)

type AI struct {
	*RocRobot
	m      					*Motion
	buttonObstacle 	*gpio.ButtonDriver
	sensorLight 		*gpio.AnalogSensorDriver
	pending					bool
	firstTime				bool
	lock chan 			bool
}

func NewAI(r *roc.Roc) *AI {

	ai := &AI{RocRobot: roc.NewRocRobot(nil)}
	ai.lock = r.AiLock
	work := func() {
	}
	gobot.On(r.Robot("motion").Event("move"), func(d interface{}) {
		r.Robot("gps").Command("sim")(map[string]interface{}{"mv": d})
	})
	ai.m = NewMotion()
	ai.m.Equal(r.Robot("motion"))
	ai.buttonObstacle = gpio.NewButtonDriver(ai.m.arduino, "buttonObstacle", "13")
	ai.sensorLight = gpio.NewAnalogSensorDriver(ai.m.arduino, "sensorLight", "0")
	ai.m.Robot.AddDevice(ai.buttonObstacle)
	ai.m.Robot.AddDevice(ai.sensorLight)
	ai.Robot = gobot.NewRobot("ai", work)
	ai.AddFunc(nil, 0, ai.pushButton, "pushButton")
	ai.AddFunc(nil, 0, ai.releaseButton, "releaseButton")
	ai.AddFunc(nil, 0, ai.pushLightButton, "pushLightButton")
	ai.obstacle()
	ai.pending = false
	ai.firstTime = true
	r.AddRocRobot(ai.RocRobot)
	return ai
}

func (ai *AI) toggle(b bool) error {

	ai.lock <- b
	if b {
		return ai.sendMessageAI(rocproto.AiInfo_LOCK)
	} else {
		return ai.sendMessageAI(rocproto.AiInfo_UNLOCK)
	}

}

func (ia *AI) sendMessageAI(id rocproto.AiInfo_Codes) error {

	var err error

	p := rocproto.Prepare(uint32(id), rocproto.Packet_DATA, rocproto.Packet_CONTROL_SERVER, rocproto.Packet_VIDEO_CLIENT)
	if misc.CheckError(err, "Sending Ai message", false) != nil {
		return err
	}
	return ia.Send(p)
}