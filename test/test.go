package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"net/http"
	"time"
	//"runtime"
	"github.com/gorilla/websocket"
)

var (
	origin     string
	url        string
	numClients int
	numMessage int
)

var t1 = make(map[string]int64)
var t2 = make(map[string]int64)
var serviceMu sync.Mutex
var comment_msg = `{"type":"comment", "img_id":210, "uid":7, "pos_x": 100, "pos_y": 200, "message":"just a test!"}`

func init() {
	flag.StringVar(&origin, "origin", "http://screenshot.com", "Origin header")
	//flag.StringVar(&url, "url", "ws://192.168.1.3:7788/ws?img_id=IMG_ID&uid=7&sid=05e82c61-4eb2-4efb-6bed-e03d1ab4ad0d", "Target URL.")
	flag.StringVar(&url, "url", "ws://127.0.0.1:8899/ws?img_id=IMG_ID&uid=7&sid=a7181603-3e3f-46b1-77f9-e46dfdc0dc26", "Target URL.")
	flag.IntVar(&numClients, "n", 100, "Number of clients.")
	flag.IntVar(&numMessage, "m", 1, "Number of messages per client.")
}

func gettime() int64 {
	return time.Now().UnixNano() / 1000000
}

func getkey(i int, j int) string {
	return strconv.Itoa(i) + "#" + strconv.Itoa(j)
}

func output_time() {
	//time.Sleep(10 * time.Second)
	var count int64 = 0
	var total int64 = 0
	for k, v := range t1 {
		diff := t2[k] - v
		if diff < 0 {
			fmt.Println("error, diff eq 0.", k, v, t2[k])
			continue
		}
		total = total + 1
		count = count + diff
		fmt.Println(k, v, t2[k], diff)
	}
	fmt.Println("total:", total, "count", count, "average", (count / total))
}

func client(client_num int, conn_num int, ul string, message string, wg *sync.WaitGroup) {
	defer wg.Done()

	//ws, err := websocket.Dial(ul, "", origin)
	var d = websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, _, err := d.Dial(ul, http.Header{"Origin": {origin}})
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("client Dial:", client_num, conn_num, ul)

	defer ws.Close()

	key := getkey(client_num, conn_num)

	for {
		var msg string
		_, p, err := ws.ReadMessage()
		msg = string(p)
		//err := websocket.Message.Receive(ws, &msg)

		if err == io.EOF {
			fmt.Println("Message.Receive", err)
			continue
		}

		if err != nil {
			fmt.Println("Message.Receive", err)
			ws.Close()
			return

		}
		if strings.Index(msg, "member") != -1 {
			//for i := 0; i < numMessage; i++ {
			serviceMu.Lock()
			if t1[key] > 0 {
				serviceMu.Unlock()
				continue
			}
			t1[key] = gettime()
			serviceMu.Unlock()
			
			err := ws.WriteMessage(websocket.TextMessage, []byte(message))
			//err := websocket.Message.Send(ws, message)
			if err != nil {
				fmt.Println("Message.Send", err, message)
				break
			}

			fmt.Println("client Send:", client_num, conn_num)
			//}
			//}()
		}

		if strings.Index(msg, "comment") != -1 {
			serviceMu.Lock()
			t2[key] = gettime()
			serviceMu.Unlock()

			//fmt.Println("client Receive:", client_num, conn_num, msg)
			//fmt.Println("Received %#v\n", msg)
			//ws.Close()
			return
			select {}
		}
	}
}

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	//fmt.Println(runtime.NumCPU())
	flag.Parse()
	var wg sync.WaitGroup
	st:= gettime()
	for i := 0; i < numClients; i++ {
		img_id := 10000 + i
		ul := strings.Replace(url, "IMG_ID", strconv.Itoa(img_id), 1)
		msg := strings.Replace(comment_msg, "210", strconv.Itoa(img_id), 1)
		//for j := 0; j < 10; j++ {
		wg.Add(1)
		go client(i, 0, ul, msg, &wg)
		//}
	}

	wg.Wait()
	nt:=gettime()
	output_time()
	fmt.Println("time",(nt-st))
}
