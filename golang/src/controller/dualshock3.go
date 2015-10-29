package controller

import (
	"fmt"
	"github.com/hybridgroup/gobot"
	"github.com/hybridgroup/gobot/platforms/joystick"
)

var ()

type Dualshock3 struct {
	Controller
	P func(int)
}

func (d Dualshock3) Start() {

	gbot := gobot.NewGobot()
	joystickAdaptor := joystick.NewJoystickAdaptor("ps3")
	joystick := joystick.NewJoystickDriver(joystickAdaptor,
		"ps3",
		"/run/media/pheos/Media/Workspace/GitHub/R.O.C-Embeded/R.O.C-CONTROLS/golang/src/controller/json/dualshock3.json",
	)
	work := func() {
		gobot.On(joystick.Event("square_press"), func(data interface{}) {
			fmt.Println("square_press")
		})
		gobot.On(joystick.Event("square_release"), func(data interface{}) {
			fmt.Println("square_release")
		})
		gobot.On(joystick.Event("triangle_press"), func(data interface{}) {
			fmt.Println("triangle_press")
		})
		gobot.On(joystick.Event("triangle_release"), func(data interface{}) {
			fmt.Println("triangle_release")
		})
		gobot.On(joystick.Event("left_x"), func(data interface{}) {
			fmt.Println("left_x", data)
			fmt.Println("Direction; axis X", data)
		})
		gobot.On(joystick.Event("left_y"), func(data interface{}) {
			fmt.Println("left_y", data)
		})
		gobot.On(joystick.Event("right_x"), func(data interface{}) {
			fmt.Println("right_x", data)
		})
		gobot.On(joystick.Event("right_y"), func(data interface{}) {
			fmt.Println("right_y", data)
		})
	}

	robot := gobot.NewRobot("joystickBot",
		[]gobot.Connection{joystickAdaptor},
		[]gobot.Device{joystick},
		work,
	)

	gbot.AddRobot(robot)

	gbot.Start()
}
