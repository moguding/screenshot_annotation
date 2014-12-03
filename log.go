package main

import log4 "screenshot_annotation/log4go"

func init() {
	log4.LoadConfiguration("log4go.xml")
	//flw := log4.NewConsoleLogWriter()
	//log4.AddFilter("stdout", log4.DEBUG, flw)
	//log4.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	//log4.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	//log4.Info("About that time, eh chaps?")
}

var log4d = log4.Debug
var log4i = log4.Info
var log4w = log4.Warn
var log4e = log4.Error
