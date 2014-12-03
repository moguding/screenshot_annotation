package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var logger *log.Logger

func init() {
	logfile, err := os.OpenFile("./log/debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		fmt.Printf("Error:", err.Error())
		os.Exit(-1)
	}
	//defer logfile.Close();
	multi := io.MultiWriter(logfile, os.Stdout)
	logger = log.New(multi, "", log.Ldate|log.Ltime)
}

func p(args ...interface{}) {
	fmt.Println(args...)
}

func pp(args ...interface{}) {
	logger.Println(args...)
}

func js(t interface{}) string {
	j, _ := json.Marshal(t)
	return string(j)
}

func jb(t interface{}) []byte {
	str := js(t)
	return []byte(str)
}

func gettime() int64 {
	return time.Now().UnixNano() / 1000000
}
