package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

var addr = flag.String("addr", ":9527", "http service address")
var debug = flag.String("debug", "false", "debug mode")
var task = flag.String("task", "", "background task")
var homeTempl = template.Must(template.ParseFiles("home.html"))
var testTempl = template.Must(template.ParseFiles("test.html"))
var rdPool *redis.Pool
var db *sql.DB
var config Configuration

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	homeTempl.Execute(w, r.Host)
}

func serveTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	img_id, _ := strconv.ParseInt(r.URL.Query().Get("img_id"), 10, 0)
	uid, _ := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 0)
	sid := r.URL.Query().Get("sid")

	type Result struct {
		ImgId int64
		Uid   int64
		Sid   string
		Host  string
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	testTempl.Execute(w, Result{img_id, uid, sid, r.Host})
}

func serveComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	img_id, _ := strconv.ParseInt(r.URL.Query().Get("img_id"), 10, 0)
	pos_x, _ := strconv.ParseInt(r.URL.Query().Get("pos_x"), 10, 0)
	pos_y, _ := strconv.ParseInt(r.URL.Query().Get("pos_y"), 10, 0)

	uid, _ := strconv.ParseInt(r.URL.Query().Get("uid"), 10, 0)
	sid := r.URL.Query().Get("sid")
	callback := strings.TrimSpace(r.URL.Query().Get("callback"))

	var content string

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if img_id == 0 {
		content = js(Comment{})
		fmt.Fprintln(w, content)
		return
	}

	u := checkLogin(int(uid), sid)
	if u == nil {
		content = js(Comment{})
		fmt.Fprintln(w, content)
		return
	}
	//todo
	doSaveComment()
	data := getCommentList(img_id, pos_x, pos_y, true)
	//fmt.Fprintln(w, js(data))
	out := callback + "(" + js(data) + ")"
	fmt.Fprintln(w, out)

	return
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log4e(err)
		}
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	config, _ = LoadConfig("config.json")
	flag.Parse()

	rdPool = initRedis(config.RedisHost)
	defer rdPool.Close()

	var err error
	db, err = sql.Open("mysql", config.MysqlHost)
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(24)
	defer db.Close()

	if *task == "message" {
		sendMessage()
	} else {
		go h.run()
		go saveComment()
		//go h.check()

		http.HandleFunc("/", serveHome)
		http.HandleFunc("/test", serveTest)
		http.HandleFunc("/comment", serveComment)
		http.HandleFunc("/ws", serveWs)
		err = http.ListenAndServe(*addr, nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}
}
