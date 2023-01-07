package main

import (
	"MyDebugger/src/api"
	"MyDebugger/src/utils"
	"fmt"
	"strings"
	"testing"
)

func NewClient() (*MyApi.MyClient, error) {
	return MyApi.NewClient("127.0.0.1:9999")
}

func NewClientWithBreakpoint() (*MyApi.MyClient, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	err = client.CreateBreakpointByFunction("main.main", "main")
	if err != nil {
		return nil, err
	}
	err = client.Continue()
	if err != nil {
		return nil, err
	}
	return client, nil
}

func PrintBreakPoints(client *MyApi.MyClient) {
	points, err := client.ListBreakpoints()
	if err != nil {
		return
	}
	utils.PrintArrayWithDetail(points)
}

func TestListBreakPoints(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
		return
	}
	PrintBreakPoints(client)
}

func TestCreateBreakPointByFunction(t *testing.T) {
	client, err := MyApi.NewClient("127.0.0.1:9999")
	if err != nil {
		t.Fatal(err)
		return
	}
	PrintBreakPoints(client)
	err = client.CreateBreakpointByFunction("main.main", "main")
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println()
	PrintBreakPoints(client)
}

func TestCreateBreakPointByAddress(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	rip := client.Current.Rip

	address := rip + 23
	points, err := client.ListBreakpoints()
	if err != nil {
		return
	}
	utils.PrintArrayWithDetail(points)
	fmt.Println()

	err = client.CreateBreakpointByAddress(address, "main")
	if err != nil {
		t.Fatal(err)
		return
	}
	points, err = client.ListBreakpoints()
	if err != nil {
		return
	}
	utils.PrintArrayWithDetail(points)

}

func TestListRegs(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	regs, err := client.ListRegs()
	if err != nil {
		t.Fatal(err)
		return
	}
	utils.PrintArrayWithDetail(regs)
}

func TestContinue(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
		return
	}
	regs, err := client.ListRegs()
	if err != nil {
		t.Fatal(err)
		return
	}
	for _, reg := range regs {
		if reg.Name == "Rip" {
			fmt.Println("RIP ==> ", reg.Value)
		}
	}
	err = client.CreateBreakpointByFunction("main.go:10", "main1")
	if err != nil {
		t.Fatal(err)
		return
	}

	err = client.Continue()
	if err != nil {
		t.Fatal(err)
		return
	}

	regs, err = client.ListRegs()
	if err != nil {
		t.Fatal(err)
		return
	}
	for _, reg := range regs {
		if reg.Name == "Rip" {
			fmt.Println("RIP ==> ", reg.Value)
		}
	}
}

func TestListSource(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	sources, err := client.ListSource()
	if err != nil {
		return
	}
	utils.PrintArrayWithDetail(sources)
}

func TestStacktrace(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	frames := client.Stacktrace()
	utils.PrintArrayWithDetail(frames)
}

func TestReadSourceCodeFromFile(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}

	codes, err := utils.ReadSourceCodeFromFile(client.Current.FilePath, client.Current.FileLine)
	if err != nil {
		t.Fatal(err)
		return
	}
	for _, code := range codes {
		fmt.Println(code)
	}
}

func TestNext(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	err = client.Next()
	if err != nil {
		t.Fatal(err)
		return
	}

	codes, err := utils.ReadSourceCodeFromFile(client.Current.FilePath, client.Current.FileLine)
	if err != nil {
		return
	}
	utils.PrintArray(codes)
}

func TestGetStat(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Println(client.Current.Regs)
}

func TestDisassembly(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	asms, err := client.Disassembly()
	if err != nil {
		t.Fatal(err)
		return
	}
	for _, asm := range asms {
		fmt.Printf("0x%x ==> %s\n", asm.Loc.PC, asm.Text)
	}
}

func TestDisassembly2(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	rip := client.Current.Rip

	ends := rip + 0x30
	asms, err := client.Disassembly2(rip, ends)
	if err != nil {
		t.Fatal(err)
		return
	}
	for _, asm := range asms {
		fmt.Printf("0x%x ==> %s\n", asm.Loc.PC, asm.Text)
	}

}

func TestStepInstruction(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}

	asms, err := client.Disassembly()
	for _, asm := range asms {
		fmt.Printf("0x%x ==> %s\n", asm.Loc.PC, asm.Text)
	}
	if err != nil {
		return
	}

	err = client.StepInstruction()
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Println()

	asms, err = client.Disassembly()
	if err != nil {
		return
	}
	for _, asm := range asms {
		fmt.Printf("0x%x ==> %s\n", asm.Loc.PC, asm.Text)
	}
}

func TestStep(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	//err = client.CreateBreakpointByFunction("main.go:")
	//if err != nil {
	//	t.Fatal(err)
	//	return
	//}
	err = client.Step()
	if err != nil {
		t.Fatal(err)
		return
	}

	codes, err := client.ReadSourceCode()
	if err != nil {
		t.Fatal(err)
		return
	}
	utils.PrintArray(codes)

	fmt.Println()

	err = client.Step()
	if err != nil {
		t.Fatal(err)
		return
	}

	codes, err = client.ReadSourceCode()
	if err != nil {
		t.Fatal(err)
		return
	}
	utils.PrintArray(codes)

}

func TestStepOut(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	err = client.Step()
	if err != nil {
		t.Fatal(err)
		return
	}
	err = client.Step()
	if err != nil {
		t.Fatal(err)
		return
	}
	codes, err := client.ReadSourceCode()
	if err != nil {
		t.Fatal(err)
		return
	}
	utils.PrintArray(codes)

	err = client.StepOut()
	if err != nil {
		t.Fatal(err)
		return
	}
	codes, err = client.ReadSourceCode()
	if err != nil {
		t.Fatal(err)
		return
	}
	utils.PrintArray(codes)

}

func TestExamineMemory(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	rip := client.Current.Rip
	memories, err := client.ExamineMemory(rip, 0x10)
	if err != nil {
		t.Fatal(err)
		return
	}
	fmt.Printf("%v\n", memories)

}

func TestSimple(t *testing.T) {
	cmd := "quit"
	cmds := strings.Split(cmd, " ")
	for _, c := range cmds {
		fmt.Println(c)
	}
}

func TestClearBreakpointByName(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	points, err := client.ListBreakpoints()
	if err != nil {
		return
	}
	utils.PrintArray(points)
	err = client.ClearBreakpointByName("main")
	if err != nil {
		return
	}
	points, err = client.ListBreakpoints()
	if err != nil {
		return
	}
	utils.PrintArray(points)
}

func TestClearBreakpointByID(t *testing.T) {
	client, err := NewClientWithBreakpoint()
	if err != nil {
		t.Fatal(err)
		return
	}
	points, err := client.ListBreakpoints()
	if err != nil {
		t.Fatal(err)
		return
	}
	utils.PrintArray(points)
	fmt.Println()
	err = client.ClearBreakpointByID(1)
	if err != nil {
		t.Fatal(err)
		return
	}
	points, err = client.ListBreakpoints()
	if err != nil {
		t.Fatal(err)
		return
	}
	utils.PrintArray(points)
}
