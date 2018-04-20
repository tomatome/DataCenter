// DataCenter.go
package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/nkovacs/go-socket.io"
)

type Room struct {
	room   string
	member []string
}

func initRoom(name, openid string) *Room {
	c := &Room{}
	c.room = name
	c.member = make([]string, 1, 2)
	c.member[0] = openid
	return c
}

var num int = 0

func getRoomId() string {
	num++
	return "chat" + strconv.Itoa(num)
}

func server() *socketio.Server {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	cRoom := make(map[string]*Room)
	server.On("connection", func(so socketio.Socket) {
		addr := so.Request().RemoteAddr
		room := "chat"
		log.Println(addr+" on connection:", len(cRoom), len(so.Rooms()))
		if len(cRoom) <= 0 {
			name := getRoomId()
			log.Println("create room:", name)
			so.Join(name)
			cRoom[name] = initRoom(name, addr)
			room = name
		} else {
			hasFound := false
			for _, c := range cRoom {
				if len(c.member) == 1 {
					c.member = append(c.member, addr)
					so.Join(c.room)
					hasFound = true
					room = c.room
					log.Println("join in room: ", c.room)
					break
				}
			}
			if !hasFound {
				name := getRoomId()
				so.Join(name)
				cRoom[name] = initRoom(name, addr)
				room = name
				log.Println("create room:", name)
			}
		}
		log.Println(so.Id())
		so.On("chat message", func(msg string) {
			log.Println("emit:<", msg, "> in room:", room)
			so.BroadcastTo(room, "chat message", addr+"说: "+msg)
			so.Emit("chat message", addr+"说: "+msg)
		})
		so.On("disconnection", func() {
			log.Println(addr + " on disconnect")
			so.BroadcastTo(room, "chat message", addr+" leaving")
			delete(cRoom, room)
		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})
	return server
}

func main() {
	server := server()
	s, _ := exec.LookPath(os.Args[0])
	log.Println("cwd:", s)
	log.Println(os.Environ())
	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("../pages/")))
	port := os.Getenv("PORT")
	host := os.Getenv("HOST")
	log.Println("Serving at ", host+":", port)
	log.Fatal(http.ListenAndServe(host+":"+port, nil))
}
