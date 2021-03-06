// DataCenter.go
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/nkovacs/go-socket.io"
)

type Room struct {
	room   string
	member int
}

func initRoom(name, openid string) *Room {
	c := &Room{}
	c.room = name
	c.member = 1
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
				if c.member == 1 {
					c.member++
					so.Join(c.room)
					hasFound = true
					room = c.room
					log.Println("join in room: ", c.room)
					so.BroadcastTo(room, "chat message", addr+" join in room")
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

		so.On("chat message", func(msg string) {
			log.Println("emit:<", msg, "> in room:", room)
			so.BroadcastTo(room, "chat message", addr+"说: "+msg)
			so.Emit("chat message", addr+"说: "+msg)
		})
		so.On("disconnection", func() {
			log.Println(addr + " on disconnect")
			so.BroadcastTo(room, "chat message", addr+" leaving")
			mRoom := cRoom[room]
			mRoom.member--
			if mRoom.member == 0 {
				delete(cRoom, room)
			}

		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})
	return server
}

func main() {
	server := server()
	log.Println(os.Environ())
	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("pages/")))
	port := os.Getenv("PORT")
	//port := "12345"
	log.Println("Serving at localhost:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
