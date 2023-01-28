package UI

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

type handler func() error

type viewInfo struct {
	title    string
	handle   handler
	view     *tview.TextView
	data     []string
	row, col int
}

// setTextView 设置 view 的内容和标题
func (info *viewInfo) setTextView() error {
	info.updateView(info.title, strings.Join(info.data, "\n"))
	return nil
}

// updateView 更新 view 的信息，包括内容和标题
func (info *viewInfo) updateView(title string, data string) {
	info.view.Clear()
	info.view.SetTitle(title)
	info.view.SetText(data)
}

// Disassembly 反汇编 Rip 附近的数据，并格式化
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
	// 对发生变化的寄存器标红
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
}

func (info *viewInfo) ExamineStack() error {
	start := client.Current.Rsp
	return info.ExamineMemory(start, 8, "hex")
}

func (info *viewInfo) StackInfo() error {
	stackFrames := client.Stacktrace()
	info.data = StacktraceToStrings(stackFrames)
	return nil
}

func (info *viewInfo) HistoryInfo() error {
	info.data = history
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
}

func (info *viewInfo) PrintAddress(addr uint64, size int) error {
	data, err := client.GetDataFromAddress(addr, size)
	if err != nil {
		return err
	}
	info.data = []string{data}
	return nil
}

func (info *viewInfo) MonitorAddress() error {
	info.data = monitors.getMonitorsData()
	return nil
}

func (info *viewInfo) TrackerAddress() error {
	info.data = trackers.getTrackersData()
	return nil
}

func NewTextViewInfo(title string, row, col int, handler func(), doneFunc func(key tcell.Key)) *viewInfo {
	info := new(viewInfo)
	info.view = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).SetChangedFunc(handler)
	info.view.SetTitle(title)
	info.view.SetBorder(true)
	info.view.SetDoneFunc(doneFunc)
	info.row = row
	info.col = col
	return info
}
