package main

import "fyne.io/fyne/v2/app"

//func main() {
//	a := app.New()
//	w := a.NewWindow("Hello World")
//
//	w.SetContent(widget.NewLabel("Hello World!"))
//	w.ShowAndRun()
//}

func main() {
	a := app.New()

	w := a.NewWindow("debugger")

	w.ShowAndRun()
}
