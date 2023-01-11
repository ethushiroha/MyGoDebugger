package main

import (
	"fmt"
	"github.com/rivo/tview"
	"strings"
)

type handler func() error

type viewInfo struct {
	//mode   string
	title  string
	handle handler
	view   *tview.TextView
	data   []string
}

// setTextView 设置 view 的内容和标题
func (info *viewInfo) setTextView() error {
	info.view.Clear()
	info.view.SetTitle(info.title)
	info.view.SetText(StringsToString(info.data))
	return nil
}

// updateView 更新 view 的信息，包括内容和标题
func (info *viewInfo) updateView(title string, data string) {
	info.view.Clear()
	info.view.SetTitle(title)
	info.view.SetText(data)
}

func (info *viewInfo) Disassembly() error {
	asms, err := client.Disassembly()
	if err != nil {
		return err
	}
	info.data = FormatASM(asms, client.Current.Rip)
	return nil
}

func (info *viewInfo) Registers() error {
	regs, err := client.ListRegs()
	if err != nil {
		return err
	}
	info.data = RegsToStrings(regs)
	oldRegs := strings.Split(info.view.GetText(false), "\n")
	if len(oldRegs) > 1 {
		for i := 1; i < 17; i++ {
			oldRegs[i] = strings.Replace(oldRegs[i], "[red]", "", -1)
			oldRegs[i] = strings.Replace(oldRegs[i], "[white]", "", -1)
			if oldRegs[i] != info.data[i] {
				info.data[i] = fmt.Sprintf("[red]%s[white]", info.data[i])
			}
		}
	}
	return nil
}

func (info *viewInfo) ExamineMemory(start, mode uint64, format string) error {
	ends := start + 0x80
	mems, err := client.ExamineMemory(start, int(ends-start))
	if err != nil {
		return err
	}
	info.data = FormatMemory(mems, start, mode, format)
	return nil
	//data := FormatMemory(mems, start, mode, format)
	//return &data, nil
}

func (info *viewInfo) ExamineStack() error {
	start := client.Current.Rsp
	return info.ExamineMemory(start, 8, "hex")
}

func (info *viewInfo) StackInfo() error {
	stackFrames := client.Stacktrace()
	info.data = StacktraceToStrings(stackFrames)
	return nil
	//data := StacktraceToStrings(stackFrames)
	//return &data, nil
}

func (info *viewInfo) HistoryInfo() error {
	length := len(history)
	if length > 8 {
		info.data = history[length-8:]
	} else {
		info.data = history
	}
	return nil
}

func (info *viewInfo) ListBreakpoints() error {
	breakpoints, err := client.ListBreakpoints()
	if err != nil {
		return err
	}
	info.data = BreakpointsToStrings(breakpoints)
	return nil
}

func (info *viewInfo) DisassemblyAddress(addr uint64) error {
	ends := addr + 0x100
	asms, err := client.Disassembly2(addr, ends)
	if err != nil {
		return err
	}
	info.data = FormatASM(asms, client.Current.Rip)
	return nil
	//data := FormatASM(asms, client.Current.Rip)
	//return &data, nil
}

func newTextView(title string, handler func()) *tview.TextView {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).SetChangedFunc(handler)
	view.SetTitle(title)
	view.SetBorder(true)
	return view
}
