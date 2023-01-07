package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strconv"
	"strings"
)

type UI struct {
	app             *tview.Application
	cmdLine         *tview.InputField
	disassemblyView *tview.TextView
	memoryView      *tview.TextView
	regsView        *tview.TextView
	grid            *tview.Grid
	data            *CurrentData
}

func (ui *UI) Run() error {
	err := ui.SetMemoryView(ui.data.MemoryData)
	if err != nil {
		return err
	}
	err = ui.SetRegsView(ui.data.RegsData)
	if err != nil {
		return err
	}
	err = ui.SetDisassemblyView(ui.data.DisassemblyData)
	if err != nil {
		return err
	}
	if err = ui.app.SetRoot(ui.grid, true).SetFocus(ui.grid).Run(); err != nil {
		return err
	}
	return nil
}

func (ui *UI) SetDisassemblyView(data []string) error {
	ui.disassemblyView.Clear()

	for _, d := range data {
		_, err := fmt.Fprintf(ui.disassemblyView, "%s\n", d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UI) SetMemoryView(data []string) error {
	ui.memoryView.Clear()

	for _, d := range data {
		_, err := fmt.Fprintf(ui.memoryView, "%s\n", d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UI) SetRegsView(data []string) error {
	ui.regsView.Clear()

	for _, d := range data {
		_, err := fmt.Fprintf(ui.regsView, "%s\n", d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ui *UI) FlashData() error {
	// err := ui.data.FlashData()
	// if err != nil {
	// 	return err
	// }
	err := ui.SetDisassemblyView(ui.data.DisassemblyData)
	if err != nil {
		return err
	}
	err = ui.SetMemoryView(ui.data.MemoryData)
	if err != nil {
		return err
	}
	err = ui.SetRegsView(ui.data.RegsData)
	if err != nil {
		return err
	}
	return nil
}

func (ui *UI) dealWithEnter(command string) error {
	var err error = nil
	tmp := strings.Split(command, " ")
	switch tmp[0] {
	case "q", "quit":
		ui.app.Stop()
		return nil
	case "b", "break":
		var name string
		if len(tmp) == 2 {
			name = ""
		} else if len(tmp) == 3 {
			name = tmp[2]
		} else {
			break
		}

		if strings.HasPrefix(tmp[1], "0x") {
			loc, err := stringToUint64(tmp[1])
			if err != nil {
				return err
			}
			err = ui.data.Client.CreateBreakpointByAddress(loc, name)
			if err != nil {
				return err
			}
		} else {
			err = ui.data.Client.CreateBreakpointByFunction(tmp[1], name)
			if err != nil {
				return err
			}
		}
		err = ui.data.Disassembly()
		if err != nil {
			return err
		}
	case "c", "continue":
		err = ui.data.Client.Continue()
		if err != nil {
			return err
		}
		err = ui.data.FlashData()
		if err != nil {
			return err
		}
	case "si", "step-instruction":
		err = ui.data.Client.StepInstruction()
		if err != nil {
			return err
		}
		err = ui.data.FlashData()
		if err != nil {
			return err
		}
	case "n", "next":
		err = ui.data.Client.Next()
		if err != nil {
			return err
		}
		err = ui.data.FlashData()
		if err != nil {
			return err
		}
	case "so", "step-out":
		err = ui.data.Client.StepOut()
		if err != nil {
			return err
		}
		err = ui.data.FlashData()
		if err != nil {
			return err
		}
	case "clear":
		if len(tmp) == 2 {
			n, err := strconv.Atoi(tmp[1])
			// is not id
			if err != nil {
				addr, err := stringToUint64(tmp[1])
				if err != nil {
					err = ui.data.Client.ClearBreakpointByName(tmp[1])
					if err != nil {
						return err
					}
				} else {
					err = ui.data.Client.ClearBreakpointByAddress(addr)
					if err != nil {
						return err
					}
				}
			} else {
				err = ui.data.Client.ClearBreakpointByID(n)
				if err != nil {
					return err
				}
			}
			err = ui.data.Disassembly()
			if err != nil {
				return err
			}
		}
	case "clear-all":
		err = ui.data.Client.ClearAllBreakpoints()
		if err != nil {
			return err
		}
	case "x":
		var mode uint64
		var format string
		var address uint64
		if len(tmp) == 2 {
			mode = 1
			//count = 1
			format = "hex"
			address, err = stringToUint64(tmp[1])
			if err != nil {
				return err
			}

		} else if len(tmp) == 3 {
			address, err = stringToUint64(tmp[2])
			if err != nil {
				return err
			}
			if len(tmp[1]) == 1 {
				mode = 1
				format = getFormat(tmp[1][0])
			} else if len(tmp[1]) == 2 {
				mode = getMode(tmp[1][0])
				format = getFormat(tmp[1][1])
			} else {
				// todo: return help error
				return err
			}
		} else {
			// todo: return help error
			return err
		}

		err = ui.data.ExamineMemory(address, mode, format)
		if err != nil {
			return err
		}

	}
	return err
}

func (ui *UI) DealWithCommand(key tcell.Key) {
	if key == tcell.KeyEnter {
		cmd := ui.cmdLine.GetText()
		err := ui.dealWithEnter(cmd)
		if err != nil {
			panic(err)
			return
		}
		ui.cmdLine.SetText("")
		err = ui.FlashData()
		if err != nil {
			return
		}
	}
}

func InitUI(address string) (*UI, error) {
	ui := new(UI)
	var err error
	ui.data, err = InitData(address)
	if err != nil {
		return nil, err
	}
	ui.app = tview.NewApplication()
	ui.cmdLine = tview.NewInputField().
		SetLabel("输入命令: ").
		SetDoneFunc(ui.DealWithCommand)

	ui.disassemblyView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})
	ui.disassemblyView.SetTitle("反汇编")
	ui.disassemblyView.SetBorder(true)

	ui.memoryView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})
	ui.memoryView.SetTitle("内存")
	ui.memoryView.SetBorder(true)

	ui.regsView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})
	ui.regsView.SetTitle("寄存器")
	ui.regsView.SetBorder(true)

	ui.grid = tview.NewGrid().
		SetRows(-20, -10, -1).
		SetColumns(-7, -3).
		SetBorders(true).
		AddItem(ui.disassemblyView, 0, 0, 1, 1, 0, 0, false).
		AddItem(ui.regsView, 0, 1, 1, 1, 0, 0, false).
		AddItem(ui.memoryView, 1, 0, 1, 1, 0, 0, false).
		AddItem(ui.cmdLine, 2, 0, 1, 2, 0, 0, true)

	return ui, nil
}
