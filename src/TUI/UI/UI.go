package UI

import (
	MyApi "MyDebugger/src/api"
	"context"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/sync/errgroup"
	"strings"
)

type UI struct {
	app        *tview.Application
	cmdLine    *tview.InputField
	grid       *tview.Grid
	views      map[string]*viewInfo
	errChannel chan error
}

type CommandHandler func([]string) error

type CommandInfo struct {
	handler  CommandHandler
	helpInfo string
}

var history []string
var client *MyApi.MyClient
var Commands map[string]*CommandInfo

func (ui *UI) Run() {
	err := ui.flashUI()
	if err != nil {
		ui.errChannel <- err
	}
	if err = ui.app.SetRoot(ui.grid, true).SetFocus(ui.grid).Run(); err != nil {
		ui.errChannel <- err
	}
}

// flashData 根据各个 view 的 handle 刷新 view data
func (ui *UI) flashData() error {
	eg, _ := errgroup.WithContext(context.Background())
	for _, value := range ui.views {
		eg.Go(value.handle)
	}
	return eg.Wait()
}

// flashUI 根据各个 view 的 title 和 data， 刷新 TUI
func (ui *UI) flashUI() error {
	eg, _ := errgroup.WithContext(context.Background())
	for _, value := range ui.views {
		eg.Go(value.setTextView)
	}
	return eg.Wait()
}

// dealWithEnter 当输入回车键之后，处理输入的指令
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
				ui.errChannel <- err
			}
		} else {
			//if cmd != lastCmd {
			history = append(history, cmd)
			//}
			err := ui.dealWithEnter(cmd)
			if err != nil {
				ui.errChannel <- err
			}
		}
	}
	err := ui.flashUI()
	if err != nil {
		ui.errChannel <- err
	}
	ui.MonitorDataChanged()
	// todo: add history
}

// MonitorError 监控 error 信息，显示在 TUI 上
// 其实在 dealWithCommand 里捕获 error 后调用也行，但是想试试 channel，练手
func (ui *UI) MonitorError() {
	for err := range ui.errChannel {
		if err != nil {
			ui.ErrorView(err.Error())
		}
	}
}

// MonitorDataChanged 当监控的数据发生变化的时候，显示在 TUI 上
func (ui *UI) MonitorDataChanged() {
	if monitors.monitorAddress() {
		ui.MonitorView2()
	}
}

func initCommands(ui *UI) {
	historyCommand := &CommandInfo{
		handler:  ui.viewHistory,
		helpInfo: "history: 查看历史命令",
	}
	stacktraceCommand := &CommandInfo{
		handler:  ui.viewStacktrace,
		helpInfo: "st/stacktrace: 查看当前调用栈",
	}
	listBreakpointCommand := &CommandInfo{
		handler:  ui.listBreakpoints,
		helpInfo: "lb/list-breakpoints: 查看当前所有的断点",
	}
	disassembleCommand := &CommandInfo{
		handler:  ui.disassembly,
		helpInfo: "d/disassemble <address>: 查看 address 处的汇编",
	}
	examineMemoryCommand := &CommandInfo{
		handler:  ui.examineMemory,
		helpInfo: "x (flag) <address>: 查看 address 处的值",
	}
	runCommand := &CommandInfo{
		handler:  ui.run,
		helpInfo: "r/run: 重新开始调试程序",
	}
	clearAllCommand := &CommandInfo{
		handler:  ui.clear,
		helpInfo: "clear-all: 清除所有断点",
	}
	clearCommand := &CommandInfo{
		handler:  ui.clear,
		helpInfo: "clear <id/name>: 根据 id 或者 name 清除某个断点",
	}
	nextCommand := &CommandInfo{
		handler:  ui.next,
		helpInfo: "n/next: 单步执行，但不进入函数内（源码层面）",
	}
	nextInCommand := &CommandInfo{
		handler:  ui.nextIn,
		helpInfo: "ni/next-in: 单步执行，但不进入函数内（汇编层面）",
	}
	stepOutCommand := &CommandInfo{
		handler:  ui.stepOut,
		helpInfo: "so/step-out: 跳出当前函数",
	}
	stepInCommand := &CommandInfo{
		handler:  ui.stepIn,
		helpInfo: "si/step-in: 单步执行，进入函数内",
	}
	continueCommand := &CommandInfo{
		handler:  ui.continues,
		helpInfo: "c/continue: 执行至下一个断点处",
	}
	createBreakpointCommand := &CommandInfo{
		handler:  ui.createBreakpoint,
		helpInfo: "b/break <address> (name): 在 address 处创建一个名为 name 的断点",
	}
	quitCommand := &CommandInfo{
		handler:  ui.quit,
		helpInfo: "q/quit/exit: 退出程序",
	}
	helpCommand := &CommandInfo{
		handler:  ui.viewHelp,
		helpInfo: "h/help <command>: 查看某个指令的用法",
	}
	focusCommand := &CommandInfo{
		handler:  ui.focus,
		helpInfo: "f/focus: 跳转到某个窗口",
	}
	printCommand := &CommandInfo{
		handler:  ui.print,
		helpInfo: "p/print <address> <size>: 显示某个地址的值",
	}
	monitorCommand := &CommandInfo{
		handler:  ui.monitor,
		helpInfo: "m/monitor <address> <size>: 监视某个地址的值",
	}
	trackCommand := &CommandInfo{
		handler:  ui.track,
		helpInfo: "tracker <action> <address> [size=4]: 跟踪地址处的值",
	}

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
		"f":                focusCommand,
		"focus":            focusCommand,
		"ni":               nextInCommand,
		"next-in":          nextInCommand,
		"p":                printCommand,
		"print":            printCommand,
		"m":                monitorCommand,
		"monitor":          monitorCommand,
		"track":            trackCommand,
	}
	wordList = getDicKeys(Commands)
}

// InitUI 用于初始化 TUI，得到默认布局
func InitUI(address string) (*UI, error) {
	var err error
	client, err = MyApi.NewClientWithMain(address)
	if err != nil {
		return nil, err
	}
	ui := new(UI)
	initCommands(ui)
	ui.errChannel = make(chan error)

	ui.views = make(map[string]*viewInfo)

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
	focusHandler := func(key tcell.Key) {
		if key == tcell.KeyEnter || key == tcell.KeyEscape {
			ui.app.SetFocus(ui.cmdLine)
		}
	}

	ui.views = map[string]*viewInfo{
		"first":  NewTextViewInfo("反汇编", 0, 0, handler, focusHandler),
		"second": NewTextViewInfo("寄存器", 0, 1, handler, focusHandler),
		"third":  NewTextViewInfo("内存", 1, 0, handler, focusHandler),
		"fourth": NewTextViewInfo("调用栈", 1, 1, handler, focusHandler),
	}

	ui.DisassemblyView()
	ui.RegistersView()
	ui.MemoryView()
	ui.StackView()

	ui.grid = tview.NewGrid().
		SetRows(21, 11, 2).
		SetColumns(-7, -3).
		SetBorders(true).
		AddItem(ui.views["first"].view, 0, 0, 1, 1, 0, 0, false).
		AddItem(ui.views["second"].view, 0, 1, 1, 1, 0, 0, false).
		AddItem(ui.views["third"].view, 1, 0, 1, 1, 0, 0, false).
		AddItem(ui.views["fourth"].view, 1, 1, 1, 1, 0, 0, false).
		AddItem(ui.cmdLine, 2, 0, 1, 2, 0, 0, true)

	err = ui.focusTo("")
	if err != nil {
		return nil, err
	}
	return ui, ui.flashData()
}
