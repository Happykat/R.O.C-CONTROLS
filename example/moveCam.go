package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Happykat/R.O.C-CONTROLS"
	"github.com/golang/protobuf/proto"
)

func main() {

	var err error
	var buff [128]byte

	conn, err := net.Dial("tcp", ":8001")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer conn.Close()

	p := roc.Gyro{X: 120, Y: 30}
	r := roc.Prepare(roc.CAM, roc.Packet_COMMAND, roc.Packet_VIDEO_CLIENT, roc.Packet_CONTROL_SERVER)
	r.Payload, err = roc.PackAny(&p)
	if err != nil {
		log.Fatal(err.Error())
	}
	b, err := proto.Marshal(r)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Sending")
	conn.Write(b)
	<-time.After(time.Second)
	i, err := conn.Read(buff[0:])
	if err != nil {
		fmt.Println(err)
		return
	}
	proto.Unmarshal(buff[:i], r)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("p: %+v", *r)
	err = roc.UnpackAny(r.Payload, &p)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("pos %v - %v\n", p.X, p.Y)
}
