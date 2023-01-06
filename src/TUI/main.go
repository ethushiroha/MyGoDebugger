package main

func main() {
	ui, err := InitUI()
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
