package main

import (
	"MyDebugger/src/TUI/UI"
	"log"
)

func main() {
	ui, err := UI.InitUI("127.0.0.1:9999")
	if err != nil {
		log.Fatal(err)
		return
	}
	go ui.MonitorError()
	ui.Run()
}
