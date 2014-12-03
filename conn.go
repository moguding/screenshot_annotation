package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/nu7hatch/gouuid"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	//"sync"
	//"os"
	//"fmt"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		if *debug == "true" {
			return true
		}
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return true
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}
		return (u.Host == r.Host || u.Host == "screenshot.com" || u.Host == "www.screenshot.com" || u.Host == "127.0.0.1:9999")
	},
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	uid    int64
	img_id int64
	sid    string
	time   int64
	addr   net.Addr
	//mu     sync.Mutex
}

// parse message.
func (c *connection) parseMessage(message []byte) *Request {
	var req Request
	e := json.Unmarshal(message, &req)
	if e != nil {
		log4w("parseMessage error,", e, string(message))
	}

	return &req
}

// check message.
func (c *connection) checkMessage(req *Request) *Err {
	req.Message = strings.TrimSpace(req.Message)
	if req.Uid == 0 || req.ImgId == 0 {
		e := &Err{Type: "error", ErrorCode: STATUS_PARAMTERS_ERROR}
		return e
	}
	if req.Type != "comment" && req.Type != "delete" && req.Type != "color" && req.Type != "create_point" {
		e := &Err{Type: "error", ErrorCode: STATUS_MESSAGE_TYPE_ERROR}
		return e
	}
	if (req.Type == "comment" || req.Type == "create_point") && req.Message == "" {
		e := &Err{Type: "error", ErrorCode: STATUS_INVAILD_MESSAGE_ERROR}
		return e
	}

	if (req.Type == "comment") && req.Color != "" {
		e := &Err{Type: "error", ErrorCode: STATUS_INVAILD_MESSAGE_ERROR}
		return e
	}

	if req.Type == "delete" && req.UniqId == "" {
		e := &Err{Type: "error", ErrorCode: STATUS_INVAILD_MESSAGE_ERROR}
		return e
	}
	if (req.Type == "color" || req.Type == "create_point") && req.Color == "" {
		e := &Err{Type: "error", ErrorCode: STATUS_INVAILD_MESSAGE_ERROR}
		return e
	}

	return nil
}

//response message
func (c *connection) getResponseMessage(t int64, uniqid string, req *Request) (res string, hasError bool) {
	err := c.checkMessage(req)
	imd_id := strconv.Itoa(req.ImgId)
	if err != nil {
		return js(err), true
	}

	var data interface{}
	if req.Type == "comment" {
		data = Response{
			Type:    "comment",
			UniqId:  uniqid,
			ImgId:   req.ImgId,
			Uid:     req.Uid,
			PosX:    req.PosX,
			PosY:    req.PosY,
			Message: req.Message,
			Time:    t,
		}
	} else if req.Type == "create_point" {
		data = CreatePoint{
			Type:    "create_point",
			UniqId:  uniqid,
			ImgId:   req.ImgId,
			Uid:     req.Uid,
			PosX:    req.PosX,
			PosY:    req.PosY,
			Message: req.Message,
			Time:    t,
			Color:   req.Color,
		}
	} else if req.Type == "delete" {
		data = Delete{
			Type:   "delete",
			UniqId: req.UniqId,
			ImgId:  req.ImgId,
			Uid:    req.Uid,
			PosX:   req.PosX,
			PosY:   req.PosY,
		}
	} else if req.Type == "color" {
		data = Color{
			Type:  "color",
			Color: req.Color,
			ImgId: req.ImgId,
			Uid:   req.Uid,
			PosX:  req.PosX,
			PosY:  req.PosY,
		}
	}

	return imd_id + SEP + js(data), false
}

// readPump pumps messages from the websocket connection to the hub.
func (c *connection) readPump(img_id, uid int64) {
	defer func() {
		//log4d("unregister in", c.ws.RemoteAddr())
		h.unregister <- c
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log4e("ReadMessage error,", c.addr.String(), err)
			break
		}

		req := c.parseMessage(message)
		if req.ImgId != int(img_id) || req.Uid != int(uid) {
			out := Err{Type: "error", ErrorCode: STATUS_MESSAGE_PARAMTERS_ERROR}
			c.write(websocket.TextMessage, jb(out))
		} else {
			u, _ := uuid.NewV4()
			key := u.String()
			t := time.Now().Unix()

			ra := rand.New(rand.NewSource(time.Now().UnixNano()))
			ts := strconv.FormatInt(time.Now().UnixNano()-ra.Int63n(10000000), 10)

			//ts := strconv.FormatInt(time.Now().UnixNano(), 10)
			r := []rune(key)
			key = string(r[0:17]) + ts

			res, hasError := c.getResponseMessage(t, key, req)

			if hasError == true {
				c.write(websocket.TextMessage, []byte(res))
			} else {
				h.broadcast <- &res

				//save
				if req.Type == "comment" || req.Type == "create_point" {
					rd := rdPool.Get()
					//RedisExec(rd, "SELECT", "3")
					json_string := JsonString{*req, t, key}
					hash_key := strings.Replace(REDIS_SCREENSHOT_COMMENT_LIST_KEY, "IMGID", strconv.FormatInt(img_id, 10), 1)
					RedisExec(rd, "ZADD", REDIS_SCREENSHOT_COMMENT_ZSET_KEY, img_id, key)
					RedisExec(rd, "HSET", hash_key, key, js(json_string))
					rd.Close()

					//f(key)
					//if req.Type == "create_point" {
					//doSaveComment()
					//}
				} else if req.Type == "delete" {
					doSaveComment()
					deleteComment(req.UniqId)
				} else if req.Type == "color" {
					doSaveComment()
					updateCommentMeta(req.ImgId, req.PosX, req.PosY, req.Color)
				}
			}
		}
	}
}

// write writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// writePump pumps messages from the hub to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		//log4d("unregister in2", c.ws.RemoteAddr())
		//h.unregister <- c
		//h.close(c)
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serverWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log4e("upgrader.Upgrade", err)
		return
	}
	//log4d("Upgrade", ws.RemoteAddr())
	img_id, err := strconv.ParseInt(r.FormValue("img_id"), 10, 0)
	if err != nil {
		log4e(err, img_id)
		return
	}
	//log4d("img_id:", img_id)

	uid, err := strconv.ParseInt(r.FormValue("uid"), 10, 0)
	if err != nil {
		log4e(err, uid)
		return
	}
	//log4d("uid:", uid)

	sid := strings.TrimSpace(r.FormValue("sid"))
	if err != nil {
		log4e(err, sid)
		return
	}
	//log4d("sid:", sid)

	c := &connection{send: make(chan []byte, 128), ws: ws, img_id: img_id, uid: uid, sid: sid, time: time.Now().Unix(), addr: ws.RemoteAddr()}
	u := checkLogin(int(uid), sid)
	if u == nil {
		out := Err{Type: "error", ErrorCode: STATUS_LOGIN_ERROR}
		log4d("no login uid:", uid, img_id, sid)
		c.write(websocket.TextMessage, jb(out))
		return
	} else {
		//todo
		log4d("u!= nil:", sid)
	}

	//log4d(h.all_connections[img_id])
	//log4d("addr:", ws.RemoteAddr())
	//log4d("register in", ws.RemoteAddr())
	h.register <- c

	var icon string
	if u.Icon == "" {
		icon = ""
	} else {
		icon = config.IconUrl + strconv.Itoa(u.Uid)
	}
	user := &User{Uid: u.Uid, Fullname: u.Fullname, Mail: u.Mail, Status: u.Status, Icon: icon}
	out := MemberInfo{Type: "member", User: []*User{user}}
	message := r.FormValue("img_id") + SEP + js(out)
	h.broadcast <- &message

	go c.writePump()
	go c.readPump(img_id, uid)
}

func memberLogin(c *connection) {
	var members []*User
	var userlist = make(map[int64]bool)
	var img_id = c.img_id

	for _, v := range h.all_connections[img_id] {
		if v == nil {
			continue
		}
		if v.addr.String() == c.addr.String() {
			continue
		}
		conn_uid := v.uid
		if userlist[conn_uid] == true {
			continue
		} else {
			userlist[conn_uid] = true
		}

		u := getUserByUid(int(conn_uid))
		if u != nil {
			var icon string
			if u.Icon == "" {
				icon = ""
			} else {
				icon = config.IconUrl + strconv.Itoa(u.Uid)
			}
			members = append(members, &User{Uid: u.Uid, Fullname: u.Fullname, Mail: u.Mail, Status: u.Status, Icon: icon})
		}
	}

	if len(members) > 0 {
		info := MemberInfo{Type: "member", User: members}
		//c.write(websocket.TextMessage, jb(info))
		select {
		case c.send <- jb(info):
		default:
			h.unregister <- c
		}
	}
}
