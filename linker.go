package roc

import (
	"errors"
	"log"
	"net"

	"net/http"

	"github.com/Happykat/R.O.C-CONTROLS/misc"
	"github.com/Happykat/R.O.C-CONTROLS/rocproto"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

// cc
type Linker struct {
	local, remote Link
	lIP, rIP      string
	lT, rT        bool
}

// Link...
//TODO clear TCP
type Link struct {
	conn    *net.TCPConn
	ws      *websocket.Conn
	out, in chan *rocproto.Packet
}

// cc
func NewLinker(lS, rS string, lT, rT bool) *Linker {

	l := Linker{local: Link{conn: nil, ws: nil, out: make(chan *rocproto.Packet, 100), in: make(chan *rocproto.Packet, 100)},
		remote: Link{conn: nil, ws: nil, out: make(chan *rocproto.Packet, 100), in: make(chan *rocproto.Packet, 100)},
		lIP:    lS, lT: lT, rIP: rS, rT: rT}
	return &l
}

func (l *Linker) Start() {
	if l.lIP != "" {
		log.Print("Staring local work")
		go l.local.startConnTCP(l.lIP, l.lT, &l.remote, rocproto.Packet_CONTROL_SERVER|rocproto.Packet_VIDEO_SERVER)
	}
	log.Println("Starting remote work")
	go l.startConnWS(l.rIP, l.rT, &l.local, rocproto.Packet_CONTROL_SERVER|rocproto.Packet_VIDEO_CLIENT)
}

//TODO timeout connection and try
func (l *Link) startConnTCP(s string, m bool, o *Link, t rocproto.Packet_Section) {

	var listener *net.TCPListener

	log.Println("Starting connection on ", s)
	tcpAddr, err := net.ResolveTCPAddr("tcp", s)
	misc.CheckError(err, "resolving address in linker.go/startConn", true)
	for {
		if m {
			log.Println("Listening on", tcpAddr.String())
			listener, err = net.ListenTCP("tcp", tcpAddr)
			misc.CheckError(err, "listening in linker.go/startConn", true)
			log.Println("Looking for a client...")
			l.conn, err = listener.AcceptTCP()
			misc.CheckError(err, "Accepting client in linker.go/startCnn", true)
			log.Print("Connection acepted")
			listener.Close()
		} else {
			log.Print("Dialing...")
			l.conn, err = net.DialTCP("tcp", nil, tcpAddr)
			misc.CheckError(err, "Dialing adresse in linker.go/startConn", true)
		}
		l.handleConn(o, t)
		log.Println("Closing connection")
		listener.Close()
		l.conn.Close()
		l.conn = nil
	}
}

//TODO Insert buffer len and check
func (l *Link) handleConn(o *Link, t rocproto.Packet_Section) {

	l.in = make(chan *rocproto.Packet, 100)
	l.out = make(chan *rocproto.Packet, 100)

	defer close(l.in)
	defer close(l.out)

	buff := make([]byte, 128)
	quit := make(chan bool)
	go func() {

		defer func() { quit <- true }()

		m := new(rocproto.Packet)
		for {
			r, err := l.conn.Read(buff[0:])
			if misc.CheckError(err, "Receiving data from conn", false) != nil {
				return
			}
			if err = checkBuffer(r, buff, m); err != nil {
				continue
			}
			routPacket(m, l, o, t)
		}
	}()
	for {
		select {
		case <-quit:
			return
		case m := <-l.out:
			b, err := proto.Marshal(m)
			if misc.CheckError(err, "linker.go/handleConn", false) != nil {
				continue
			}
			_, err = l.conn.Write(b)
			if misc.CheckError(err, "linker.go/handleConn", false) != nil {
				return
			}
		}
	}
}

func (l *Linker) startConnWS(s string, m bool, o *Link, t rocproto.Packet_Section) {

	log.Printf("Starting websocket on %v/controls", s)
	if m {
		http.HandleFunc("/controls", l.listenRemoteWS)
		err := http.ListenAndServe(s, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	} else {
		log.Fatalln("Websocket client not yet implemented")
	}
}

func (l *Linker) listenRemoteWS(w http.ResponseWriter, r *http.Request) {

	if l.remote.ws != nil {
		er := "Remote connection already taken !!"
		log.Println(er)
		w.Write([]byte(er))
		return
	}
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		er := "Error upgrading to websocket"
		log.Println(er)
		w.Write([]byte(er))
		return
	}
	log.Println("Client connected", r.RemoteAddr)
	l.remote.ws = c
	go l.remote.handleWS(&l.local, rocproto.Packet_CONTROL_SERVER|rocproto.Packet_VIDEO_CLIENT)
	return
}

func (l *Linker) listenLocalWS(w http.ResponseWriter, r *http.Request) {

	if l.local.ws != nil {
		er := "Local connection already taken !!"
		log.Println(er)
		w.Write([]byte(er))
	}
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		er := "Error upgrading to websocket"
		log.Println(er)
		w.Write([]byte(er))
		return
	}
	log.Println("Client connected", r.RemoteAddr)
	l.local.ws = c
	go l.local.handleWS(&l.remote, rocproto.Packet_CONTROL_SERVER|rocproto.Packet_VIDEO_SERVER)
	return
}

func (l *Link) handleWS(o *Link, t rocproto.Packet_Section) {

	l.in = make(chan *rocproto.Packet, 100)
	l.out = make(chan *rocproto.Packet, 100)

	defer l.ws.Close()
	defer func() { l.ws = nil }()
	defer close(l.in)
	defer close(l.out)

	quit := make(chan bool)
	go func() {

		defer func() { quit <- true }()

		m := new(rocproto.Packet)
		for {
			_, buff, err := l.ws.ReadMessage()
			log.Println("Received ==>	", buff)
			if misc.CheckError(err, "Receiving data from conn", false) != nil {
				return
			}
			if err = checkBuffer(-1, buff, m); err != nil {
				//TODO error notification
				continue
			}
			routPacket(m, l, o, t)
		}
	}()
	for {
		select {
		case <-quit:
			return
		case m := <-l.out:
			b, err := proto.Marshal(m)
			if misc.CheckError(err, "linker.go/handleWS", false) != nil {
				continue
			}
			err = l.ws.WriteMessage(websocket.BinaryMessage, b)
			if misc.CheckError(err, "linker.go/handleWS", false) != nil {
				return
			}
		}
	}
}

func checkBuffer(r int, buff []byte, m *rocproto.Packet) (err error) {

	if r > -1 {
		err = proto.Unmarshal(buff[0:r], m)
	} else {
		err = proto.Unmarshal(buff[0:], m)
	}
	if err != nil {
		log.Println("Cannot Unmarshall packet", err.Error())
		log.Println(r, buff[0:])
		return err
	}
	if m.Magic != MAGIC {
		log.Println("Wrong packet")
		return errors.New("Wrong packet")
	}
	return nil
}

func routPacket(m *rocproto.Packet, l, o *Link, t rocproto.Packet_Section) {
	if m.Header&uint32(t) != 0 {
		l.in <- m
	}
	if (m.Header&uint32(rocproto.Packet_MASK_DEST))&^uint32(t) != 0 && o.conn != nil {
		o.out <- m
	}
}

// cc
func (l *Linker) Stop() {
	if l.remote.conn != nil {
		l.remote.conn.Close()
		close(l.remote.in)
		close(l.remote.out)
	}
	if l.local.conn != nil {
		l.local.conn.Close()
		close(l.local.in)
		close(l.local.out)
	}
}

//cc
func (l *Linker) Send(p *rocproto.Packet) error {

	if (p.Header&uint32(rocproto.Packet_VIDEO_CLIENT)) != 0 && l.remote.conn != nil {
		l.remote.out <- p
	}
	if (p.Header&uint32(rocproto.Packet_VIDEO_SERVER)) != 0 && l.local.conn != nil {
		if l.local.conn != nil {
			l.local.out <- p
		} else {
			return errors.New("Local connection not established could not send message")
		}
	}
	return nil
}

//cc
func (l *Linker) RegisterChannel(r bool) chan *rocproto.Packet {

	if r {
		return l.remote.in
	}
	return l.local.in
}
