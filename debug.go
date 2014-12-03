package main

import (
	"fmt"
	"os"
)

var datafile *os.File

func init() {
	logfile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		fmt.Printf("Error:", err.Error())
		os.Exit(-1)
	}
	datafile = logfile
}

func f(args ...interface{}) {
	fmt.Fprintln(datafile, args...)
}
