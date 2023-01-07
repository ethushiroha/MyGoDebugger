package main

func main() {
	ui, err := InitUI("127.0.0.1:9999")
	if err != nil {
		panic(err)
		return
	}
	err = ui.Run()
	if err != nil {
		panic(err)
		return
	}
}
