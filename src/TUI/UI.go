package main

import (
	MyApi "MyDebugger/src/api"
	"MyDebugger/src/utils"
	"context"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
)

type UI struct {
	app     *tview.Application
	cmdLine *tview.InputField
	//disassemblyView *tview.TextView
	//memoryView      *tview.TextView
	//regsView        *tview.TextView
	//infoView        *tview.TextView
	grid *tview.Grid
	//data        *CurrentData
	views map[string]*viewInfo
}

type CommandHandler func([]string) error

type CommandInfo struct {
	handler  CommandHandler
	helpInfo string
}

var history []string

var client *MyApi.MyClient
var Commands map[string]*CommandInfo

func (ui *UI) Run() error {
	err := ui.flashUI()
	if err != nil {
		return err
	}
	if err = ui.app.SetRoot(ui.grid, true).SetFocus(ui.grid).Run(); err != nil {
		return err
	}
	return nil
}

func (ui *UI) flashData() error {
	eg, _ := errgroup.WithContext(context.Background())
	for _, value := range ui.views {
		//var err error
		//value.data, err = value.handle()
		//if err != nil {
		//	return err
		//}
		eg.Go(value.handle)
	}
	return eg.Wait()
}

func (ui *UI) flashUI() error {
	eg, _ := errgroup.WithContext(context.Background())

	for _, value := range ui.views {
		eg.Go(value.setTextView)
	}

	return eg.Wait()
}

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
		view.title = "内存器"
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

// view ends

// command start

// quit 执行退出指令
func (ui *UI) quit(args []string) error {
	ui.app.Stop()
	return nil
}

// createBreakpoint 下断点
func (ui *UI) createBreakpoint(args []string) error {
	if args == nil || len(args) == 0 {
		return ui.viewHelp([]string{"b"})
	}
	var name string
	var err error
	if len(args) == 1 {
		name = ""
	} else if len(args) == 2 {
		name = args[1]
	} else {
		return ui.viewHelp([]string{"b"})
	}

	if strings.HasPrefix(args[0], "0x") {
		loc, err := utils.StringToUint64(args[0])
		if err != nil {
			return err
		}
		err = client.CreateBreakpointByAddress(loc, name)
		if err != nil {
			return err
		}
	} else {
		err = client.CreateBreakpointByFunction(args[0], name)
		if err != nil {
			return err
		}
	}
	ui.BreakpointsView()
	return ui.flashData()
}

// continues 执行到下一个断点处
func (ui *UI) continues(args []string) error {
	err := client.Continue()
	if err != nil {
		return err
	}
	return ui.flashData()
}

// stepIn 单步执行，进入函数内
func (ui *UI) stepIn(args []string) error {
	err := client.StepInstruction()
	if err != nil {
		return err
	}
	return ui.flashData()
}

// stepOut 跳出当前函数
func (ui *UI) stepOut(args []string) error {
	err := client.StepOut()
	if err != nil {
		return err
	}
	return ui.flashData()
}

// next 单步执行，不进入函数（源码层面）
func (ui *UI) next(args []string) error {
	err := client.Next()
	if err != nil {
		return err
	}
	return ui.flashData()
}

// clear 清除断点
func (ui *UI) clear(args []string) error {
	if args == nil || len(args) == 0 {
		return ui.viewHelp([]string{"clear"})
	}
	if len(args) == 1 {
		n, err := strconv.Atoi(args[0])
		// is not id
		if err != nil {
			addr, err := utils.StringToUint64(args[0])
			if err != nil {
				err = client.ClearBreakpointByName(args[0])
				if err != nil {
					return err
				}
			} else {
				err = client.ClearBreakpointByAddress(addr)
				if err != nil {
					return err
				}
			}
		} else {
			err = client.ClearBreakpointByID(n)
			if err != nil {
				return err
			}
		}
		if view, ok := ui.views["first"]; ok {
			view.handle = view.Disassembly
			view.title = "反汇编"
		}
		ui.BreakpointsView()
		return ui.flashData()
	} else {
		return ui.viewHelp([]string{"clear"})
	}
}

// clearAll 清除所有的断点
func (ui *UI) clearAll(args []string) error {
	err := client.ClearAllBreakpoints()
	if err != nil {
		return err
	}
	if view, ok := ui.views["first"]; ok {
		view.handle = view.Disassembly
		view.title = "反汇编"
	}
	ui.BreakpointsView()
	return ui.flashData()
}

// run 重新开始调试程序
func (ui *UI) run(args []string) error {
	err := client.ReRun(false)
	if err != nil {
		return err
	}
	return ui.flashData()
}

// examineMemory 查看从某地址开始的内存数据
func (ui *UI) examineMemory(args []string) error {
	if args == nil || len(args) == 0 {
		return ui.viewHelp([]string{"x"})
	}
	var mode uint64
	var format string
	var address uint64
	var err error
	if len(args) == 1 {
		mode = 1
		//count = 1
		format = "hex"
		address, err = utils.StringToUint64(args[0])
		if err != nil {
			return err
		}

	} else if len(args) == 2 {
		address, err = utils.StringToUint64(args[1])
		if err != nil {
			return err
		}
		if len(args[0]) == 1 {
			mode = 1
			format = getFormat(args[0][0])
		} else if len(args[0]) == 2 {
			mode = getMode(args[0][0])
			format = getFormat(args[0][1])
		} else {
			// todo: return help error
			return err
		}
	} else {
		// todo: return help error
		return err
	}

	if view, ok := ui.views["third"]; ok {
		err = view.ExamineMemory(address, mode, format)
		if err != nil {
			return err
		}
	}
	return nil
}

// disassembly 查看从某地址开始的汇编代码
func (ui *UI) disassembly(args []string) error {
	if args == nil || len(args) == 0 {
		return ui.viewHelp([]string{"d"})
	}
	if len(args) != 1 {
		// todo: return help error
		return nil
	}
	addr, err := utils.StringToUint64(args[0])
	if err != nil {
		return err
	}
	if view, ok := ui.views["first"]; ok {
		err = view.DisassemblyAddress(addr)
		if err != nil {
			return err
		}
	}
	return nil
}

// listBreakpoints 列出当前所有的断点
func (ui *UI) listBreakpoints(args []string) error {
	ui.BreakpointsView()
	return ui.flashData()
}

// viewStacktrace 查看当前的调用栈
func (ui *UI) viewStacktrace(args []string) error {
	ui.StackView()
	return ui.flashData()
}

// viewHistory 查看历史命令记录
func (ui *UI) viewHistory(args []string) error {
	ui.HistoryView()
	return ui.flashData()
}

// viewHelp 查看帮助信息
func (ui *UI) viewHelp(args []string) error {
	if args == nil || len(args) != 1 {
		return ui.viewHelp([]string{"h"})
	}
	if cmd, ok := Commands[args[0]]; ok {
		ui.HelpView(cmd.helpInfo)
	}
	return nil
}

// command ends

func (ui *UI) dealWithEnter(command string) error {
	tmp := strings.Split(command, " ")
	cmd := tmp[0]
	var args []string
	if len(tmp) > 1 {
		args = tmp[1:]
	} else {
		args = nil
	}
	if command, ok := Commands[cmd]; ok {
		return command.handler(args)
	}
	return nil
}

func (ui *UI) dealWithCommand(key tcell.Key) {
	if key == tcell.KeyEnter {
		cmd := ui.cmdLine.GetText()
		ui.cmdLine.SetText("")
		var lastCmd string
		if len(history) != 0 {
			lastCmd = history[len(history)-1]
		}
		if cmd == "" {
			history = append(history, lastCmd)
			err := ui.dealWithEnter(lastCmd)
			if err != nil {
				ui.ErrorView(err.Error())
				return
			}
		} else {
			//if cmd != lastCmd {
			history = append(history, cmd)
			//}
			err := ui.dealWithEnter(cmd)
			if err != nil {
				ui.ErrorView(err.Error())
				return
			}
		}
		err := ui.flashUI()
		if err != nil {
			ui.ErrorView(err.Error())
			return
		}
	}
	// todo: add history
}

func initCommands(ui *UI) {
	historyCommand := new(CommandInfo)
	historyCommand.handler = ui.viewHistory
	// todo: add history location
	historyCommand.helpInfo = "history: 查看历史命令"

	stacktraceCommand := new(CommandInfo)
	stacktraceCommand.handler = ui.viewStacktrace
	stacktraceCommand.helpInfo = "st/stacktrace: 查看当前调用栈"

	listBreakpointCommand := new(CommandInfo)
	listBreakpointCommand.handler = ui.listBreakpoints
	listBreakpointCommand.helpInfo = "lb/list-breakpoints: 查看当前所有的断点"

	disassembleCommand := new(CommandInfo)
	disassembleCommand.handler = ui.disassembly
	disassembleCommand.helpInfo = "d/disassemble <address>: 查看 address 处的汇编"

	examineMemoryCommand := new(CommandInfo)
	examineMemoryCommand.handler = ui.examineMemory
	examineMemoryCommand.helpInfo = "x (flag) <address>: 查看 address 处的值"

	runCommand := new(CommandInfo)
	runCommand.handler = ui.run
	runCommand.helpInfo = "r/run: 重新开始调试程序"

	clearAllCommand := new(CommandInfo)
	clearAllCommand.handler = ui.clearAll
	clearAllCommand.helpInfo = "clear-all: 清除所有断点"

	clearCommand := new(CommandInfo)
	clearCommand.handler = ui.clear
	clearCommand.helpInfo = "clear <id/name>: 根据 id 或者 name 清除某个断点"

	nextCommand := new(CommandInfo)
	nextCommand.handler = ui.next
	nextCommand.helpInfo = "next: 单步执行，但不进入函数内"

	stepOutCommand := new(CommandInfo)
	stepOutCommand.handler = ui.stepOut
	stepOutCommand.helpInfo = "so/step-out: 跳出当前函数"

	stepInCommand := new(CommandInfo)
	stepInCommand.handler = ui.stepIn
	stepInCommand.helpInfo = "si/step-in: 单步执行，进入函数内"

	continueCommand := new(CommandInfo)
	continueCommand.handler = ui.continues
	continueCommand.helpInfo = "c/continue: 执行至下一个断点处"

	createBreakpointCommand := new(CommandInfo)
	createBreakpointCommand.handler = ui.createBreakpoint
	createBreakpointCommand.helpInfo = "b/break <address> (name): 在 address 处创建一个名为 name 的断点"

	quitCommand := new(CommandInfo)
	quitCommand.handler = ui.quit
	quitCommand.helpInfo = "q/quit/exit: 退出程序"

	helpCommand := new(CommandInfo)
	helpCommand.handler = ui.viewHelp
	helpCommand.helpInfo = "h/help <command>: 查看某个指令的用法"

	Commands = map[string]*CommandInfo{
		"quit":             quitCommand,
		"q":                quitCommand,
		"exit":             quitCommand,
		"b":                createBreakpointCommand,
		"break":            createBreakpointCommand,
		"c":                continueCommand,
		"continue":         continueCommand,
		"si":               stepInCommand,
		"step-in":          stepInCommand,
		"so":               stepOutCommand,
		"step-out":         stepOutCommand,
		"n":                nextCommand,
		"next":             nextCommand,
		"clear":            clearCommand,
		"clear-all":        clearAllCommand,
		"r":                runCommand,
		"run":              runCommand,
		"x":                examineMemoryCommand,
		"d":                disassembleCommand,
		"disassemble":      disassembleCommand,
		"lb":               listBreakpointCommand,
		"list-breakpoints": listBreakpointCommand,
		"st":               stacktraceCommand,
		"stacktrace":       stacktraceCommand,
		"history":          historyCommand,
		"h":                helpCommand,
		"help":             helpCommand,
	}

	wordList = getDicKeys(Commands)
}

func InitUI(address string) (*UI, error) {
	var err error
	client, err = MyApi.NewClientWithMain(address)
	if err != nil {
		return nil, err
	}
	ui := new(UI)
	initCommands(ui)

	ui.views = make(map[string]*viewInfo)
	if err != nil {
		return nil, err
	}

	ui.app = tview.NewApplication()
	history = make([]string, 0)

	ui.cmdLine = tview.NewInputField().
		SetLabel("输入命令: ").
		SetDoneFunc(ui.dealWithCommand)
	ui.cmdLine.SetAutocompleteFunc(AutoComplete)
	ui.cmdLine.SetAutocompletedFunc(func(text string, index int, source int) bool {
		if source != tview.AutocompletedNavigate {
			ui.cmdLine.SetText(text)
		}
		return source == tview.AutocompletedEnter || source == tview.AutocompletedClick
	})

	handler := func() {
		ui.app.Draw()
	}

	view1 := new(viewInfo)
	view1.view = newTextView("反汇编", handler)

	view2 := new(viewInfo)
	view2.view = newTextView("内存", handler)

	view3 := new(viewInfo)
	view3.view = newTextView("寄存器", handler)

	view4 := new(viewInfo)
	view4.view = newTextView("调用栈", handler)

	ui.views = map[string]*viewInfo{
		"first":  view1,
		"second": view3,
		"third":  view2,
		"fourth": view4,
	}

	ui.DisassemblyView()
	ui.RegistersView()
	ui.MemoryView()
	ui.StackView()

	ui.grid = tview.NewGrid().
		SetRows(21, 11, 2).
		SetColumns(-7, -3).
		SetBorders(true).
		AddItem(view1.view, 0, 0, 1, 1, 0, 0, false).
		AddItem(view3.view, 0, 1, 1, 1, 0, 0, false).
		AddItem(view2.view, 1, 0, 1, 1, 0, 0, false).
		AddItem(view4.view, 1, 1, 1, 1, 0, 0, false).
		AddItem(ui.cmdLine, 2, 0, 1, 2, 0, 0, true)

	err = ui.flashData()
	if err != nil {
		return nil, err
	}
	return ui, nil
}
