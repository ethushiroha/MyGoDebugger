package UI

import (
	"MyDebugger/src/utils"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"reflect"
	"strings"
)

// view start

// ErrorView 是当执行指令出现错误时，会调用，在右下角的区块中显示报错信息
func (ui *UI) ErrorView(data string) {
	if view, ok := ui.views["fourth"]; ok {
		view.updateView("错误信息", data)
	}
}

// StackView 是在右下角显示当前的调用栈信息
func (ui *UI) StackView() {
	if view, ok := ui.views["fourth"]; ok {
		view.handle = view.StackInfo
		view.title = "调用栈"
	}
}

// HistoryView 是在右下角显示历史命令记录
func (ui *UI) HistoryView() {
	if view, ok := ui.views["fourth"]; ok {
		view.handle = view.HistoryInfo
		view.title = "历史记录"
	}
}

// BreakpointsView 是在右下角显示所有的断点
func (ui *UI) BreakpointsView() {
	if view, ok := ui.views["fourth"]; ok {
		view.handle = view.ListBreakpoints
		view.title = "断点"
	}
}

// RegistersView 是在右上角显示寄存器信息
func (ui *UI) RegistersView() {
	if view, ok := ui.views["second"]; ok {
		view.handle = view.Registers
		view.title = "寄存器"
	}
}

// MemoryView 是在左下角显示内存信息，默认起始地址为 Rsp 寄存器的值
func (ui *UI) MemoryView() {
	if view, ok := ui.views["third"]; ok {
		view.handle = view.ExamineStack
		view.title = "内存"
	}
}

// DisassemblyView 是在左上角显示反汇编内容的指针，默认起始地址为 Rip 寄存器的值
func (ui *UI) DisassemblyView() {
	if view, ok := ui.views["first"]; ok {
		view.handle = view.Disassembly
		view.title = "反汇编"
	}
}

// HelpView 是在右下角显示帮助信息
func (ui *UI) HelpView(info string) {
	if view, ok := ui.views["fourth"]; ok {
		view.data = []string{info}
		view.title = "帮助信息"
	}
}

// focusTo 聚焦到某个 Item
func (ui *UI) focusTo(name string) error {
	ui.grid = tview.NewGrid().
		SetRows(21, 11, 2).
		SetColumns(-7, -3).
		SetBorders(true)
	flag := false
	for viewName, view := range ui.views {
		if viewName == name {
			view.view.SetBorderColor(tcell.ColorRed)
			ui.grid.AddItem(view.view, view.row, view.col, 1, 1, 0, 0, true)
			flag = true
		} else {
			view.view.SetBorderColor(tcell.ColorWhite)
			ui.grid.AddItem(view.view, view.row, view.col, 1, 1, 0, 0, false)
		}
	}
	if flag {
		ui.grid.AddItem(ui.cmdLine, 2, 0, 1, 2, 0, 0, false)
	} else {
		ui.grid.AddItem(ui.cmdLine, 2, 0, 1, 2, 0, 0, true)
	}
	ui.app.SetFocus(ui.grid)
	return nil
}

// 另一种 focusTo 的方法，反射，感觉不是很优雅
func (ui *UI) focusTo2(name string) error {
	if name == "" {
		return nil
	}
	items := utils.GetStructPtrUnExportedField(ui.grid, "items")
	// 将全部的 focus 值全部设置为 false
	for i := 0; i < items.Len(); i++ {
		flag := items.Index(i).Elem().FieldByName("Focus")
		rv := reflect.ValueOf(false)
		flag.Set(rv)
	}
	for key, view := range ui.views {
		view.view.SetBorderColor(tcell.ColorWhite)
		// 找到需要 focus 的 view
		if key == name {
			view.view.SetBorderColor(tcell.ColorRed)
			// 设置他的 focus 值为 true
			for i := 0; i < items.Len(); i++ {
				viewAddr := fmt.Sprintf("%p", view.view)
				nAddr := fmt.Sprintf("0x%x", items.Index(i).Elem().FieldByName("Item"))
				// 两个的地址一样就是同一个对象
				if viewAddr == nAddr {
					flag := items.Index(i).Elem().FieldByName("Focus")
					rv := reflect.ValueOf(true)
					flag.Set(rv)
					break
				}
			}
			break
		}
	}
	ui.app.SetFocus(ui.grid)
	return nil
}

// MonitorView 监视内存
func (ui *UI) MonitorView() error {
	if view, ok := ui.views["fourth"]; ok {
		view.handle = view.MonitorAddress
		view.title = "监视器"
	}
	return nil
}

// MonitorView2 另一种修改 View 的方式，和 ErrorView 一样，直接用 view.updateView 来更改内容
// 当监控的数据发生变化的时候，才会调用
func (ui *UI) MonitorView2() {
	if view, ok := ui.views["fourth"]; ok {
		view.updateView("监视器", strings.Join(monitors.getMonitorsData(), "\n"))
	}
}

// view ends
