package main

import (
	//"github.com/gorilla/websocket"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex

type hub struct {
	broadcast chan *string

	register chan *connection

	unregister chan *connection

	all_connections map[int64][]*connection
}

var h = hub{
	broadcast:       make(chan *string, 512),
	register:        make(chan *connection),
	unregister:      make(chan *connection),
	all_connections: make(map[int64][]*connection),
}

const SEP = "___SEP___"

func (h *hub) check() {
	//for {
	//mu.Lock()
	for _, conn := range h.all_connections {
		for _, c := range conn {
			if c == nil {
				continue
			}

			t := time.Now().Unix()

			if t-c.time > 6000 {
				log4d("timeout,", c.addr.String())
				h.unregister <- c
			}
		}
	}
	//mu.Unlock()
	//time.Sleep(3 * time.Minute)
	//}
}

func (h *hub) close(c *connection) {
	//mu.Lock()
	//defer mu.Unlock()
	//c.write(websocket.CloseMessage, []byte{})
	close(c.send)
	c.ws.Close()
	c = nil
}

func (h *hub) run() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case c := <-h.register:
			log4d("register, true", c.addr.String())
			h.all_connections[c.img_id] = append(h.all_connections[c.img_id], c)
			memberLogin(c)
		case c := <-h.unregister:
			log4d("unregister, true", c.addr.String())
			for index, conn := range h.all_connections[c.img_id] {
				if conn == nil {
					continue
				}
				if conn.addr.String() == c.addr.String() {
					h.close(conn)
					h.all_connections[c.img_id][index] = nil
					break
				}
			}
		case m := <-h.broadcast:
			arr := strings.Split(*m, SEP)
			image_id_str := arr[0]
			message := arr[1]
			img_id, _ := strconv.ParseInt(image_id_str, 10, 0)

			for _, c := range h.all_connections[img_id] {
				if c == nil {
					continue
				}
				select {
				case c.send <- []byte(message):
				default:
					h.unregister <- c
				}
			}
		case <-ticker.C:
			h.check()
		}
	}
}
