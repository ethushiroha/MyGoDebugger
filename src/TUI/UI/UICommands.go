package UI

import (
	"MyDebugger/src/utils"
	"fmt"
	"github.com/Knetic/govaluate"
	"strconv"
	"strings"
)

// command start

// quit 执行退出指令
func (ui *UI) quit(args []string) error {
	close(ui.errChannel)
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

// nextIn 单步执行，不进入函数（汇编层面）
func (ui *UI) nextIn(args []string) error {
	err := client.NextInstruction()
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

func (ui *UI) focus(args []string) error {
	if args == nil || len(args) == 0 {
		return ui.viewHelp([]string{"f"})
	}
	if len(args) == 1 {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		switch id {
		case 1:
			return ui.focusTo("first")
		case 2:
			return ui.focusTo("second")
		case 3:
			return ui.focusTo("third")
		case 4:
			return ui.focusTo("fourth")

		}
	}
	return nil
}

func (ui *UI) print(args []string) error {
	if args == nil || len(args) == 0 {
		return nil
	}
	if len(args) != 2 {
		return nil
	}
	address, err := utils.StringToUint64(args[0])
	if err != nil {
		return err
	}
	size, err := strconv.Atoi(args[1])
	if err != nil {
		return nil
	}
	if view, ok := ui.views["fourth"]; ok {
		err = view.PrintAddress(address, size)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UI) monitor(args []string) error {
	if args == nil || len(args) == 0 {
		return ui.viewMonitors()
	}
	var size int
	var address string
	var err error
	// todo: 支持寄存器和运算
	expr, err := RegArgs(args[0])
	expression, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		return err
	}
	result, err := expression.Evaluate(nil)
	if err != nil {
		return err
	}
	address = fmt.Sprintf("0x%x", int(result.(float64)))
	if err != nil {
		return err
	}

	switch len(args) {
	case 1:
		size = 4
	case 2:
		size, err = strconv.Atoi(args[1])
		if err != nil {
			return err
		}
	default:
		return ui.viewHelp([]string{"m"})
	}
	monitors.add(address, size)
	return ui.viewMonitors()
}

func (ui *UI) viewMonitors() error {
	monitors.monitorAddress()
	err := ui.MonitorView()
	if err != nil {
		return err
	}
	return ui.flashData()
}

// command ends
