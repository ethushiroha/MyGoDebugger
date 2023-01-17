package MyApi

import (
	"MyDebugger/src/utils"
	"fmt"
	"github.com/go-delve/delve/service/api"
	"github.com/go-delve/delve/service/rpc2"
	"strings"
)

// CurrentStatus 表示运行时的确切状态
type CurrentStatus struct {
	// Statement 表示当前帧
	Statement int
	// ThreadID 表示当前线程 ID
	ThreadID int
	// GoroutineID 表示当前所在协程 ID
	GoroutineID int64
	// FilePath 表示当前源文件位置
	FilePath string
	// FileLine 表示当前源文件行号
	FileLine int
	// Regs 表示寄存器
	Regs api.Registers
	// Rip 寄存器的值，经常需要使用
	Rip uint64
	// Rsp 寄存器的值，经常需要使用
	Rsp uint64
	// Rbp 寄存器的值，经常需要使用
	Rbp uint64
}

type MyClient struct {
	// client 是调用 rpc 的客户端
	client  *rpc2.RPCClient
	Current *CurrentStatus
}

func NewClient(addr string) (*MyClient, error) {
	c := new(MyClient)
	c.client = rpc2.NewClient(addr)
	c.Current = new(CurrentStatus)
	c.Current.Regs = nil
	err := c.GetStat()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func NewClientWithMain(addr string) (*MyClient, error) {
	client, err := NewClient(addr)
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

// currentEvalScope 返回当前的状态
func (c *MyClient) currentEvalScope() api.EvalScope {
	return api.EvalScope{
		GoroutineID:  c.Current.GoroutineID,
		Frame:        c.Current.Statement,
		DeferredCall: 0,
	}
}

// CreateBreakpointByFunction 在函数名开始处下断点
func (c *MyClient) CreateBreakpointByFunction(functionName, breakpointName string) error {
	location, err := c.FindLocationByName(functionName)
	if err != nil {
		return err
	}
	_, err = c.client.CreateBreakpoint(&api.Breakpoint{
		Name: breakpointName,
		File: location.File,
		Line: location.Line,
	})
	if err != nil {
		return err
	}
	return nil
}

// CreateBreakpointByAddress 根据地址下断点
func (c *MyClient) CreateBreakpointByAddress(addr uint64, name string) error {
	_, err := c.client.CreateBreakpoint(&api.Breakpoint{
		Name: name,
		Addr: addr,
	})
	if err != nil {
		return err
	}
	return nil
}

// ListBreakpoints 列出所有断点
func (c *MyClient) ListBreakpoints() ([]*api.Breakpoint, error) {
	breakpoints, err := c.client.ListBreakpoints(true)
	if err != nil {
		return nil, err
	}
	return breakpoints, nil
}

// Continue 运行到下一个断点处
func (c *MyClient) Continue() error {
	ch := c.client.Continue()
	// 等待 continue 执行完毕
	<-ch
	return c.GetStat()
}

// ListRegs 得到所有的寄存器
func (c *MyClient) ListRegs() (api.Registers, error) {
	registers, err := c.client.ListThreadRegisters(c.Current.ThreadID, true)
	// registers, err := c.client.ListScopeRegisters(c.currentEvalScope(), true)
	if err != nil {
		return nil, err
	}
	return registers, nil
}

func (c *MyClient) ListSource() ([]string, error) {
	sources, err := c.client.ListSources("")
	if err != nil {
		return nil, err
	}
	return sources, nil
}

// FindLocationByName 根据函数名称找到位置信息
func (c *MyClient) FindLocationByName(functionName string) (api.Location, error) {
	locations, err := c.client.FindLocation(c.currentEvalScope(), functionName, true, nil)
	if err != nil {
		return api.Location{}, err
	}
	return locations[0], nil
}

// Stacktrace 列出调用栈
func (c *MyClient) Stacktrace() []api.Stackframe {
	stacktrace, err := c.client.Stacktrace(c.Current.GoroutineID, 10, api.StacktraceSimple, nil)
	if err != nil {
		return nil
	}
	return stacktrace
}

// GetStat 更新当前状态
func (c *MyClient) GetStat() error {
	state, err := c.client.GetState()
	if err != nil {
		return err
	}
	c.Current.Statement = 0

	if state.SelectedGoroutine != nil && state.SelectedGoroutine.ID > 0 {
		c.Current.GoroutineID = state.SelectedGoroutine.ID
	}

	if state.CurrentThread != nil {
		c.Current.ThreadID = state.CurrentThread.ID
		c.Current.FilePath = state.CurrentThread.File
		c.Current.FileLine = state.CurrentThread.Line
	}

	regs, err := c.ListRegs()
	if err != nil {
		return err
	}
	// 寄存器的值
	if regs != nil {
		c.Current.Regs = regs
		for _, reg := range regs {
			if reg.Name == "Rip" {
				c.Current.Rip, err = utils.StringToUint64(reg.Value)
				if err != nil {
					return err
				}
			} else if reg.Name == "Rsp" {
				c.Current.Rsp, err = utils.StringToUint64(reg.Value)
				if err != nil {
					return err
				}
			} else if reg.Name == "Rbp" {
				c.Current.Rbp, err = utils.StringToUint64(reg.Value)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Next 是步过，不会进入函数内，源码层面
func (c *MyClient) Next() error {
	_, err := c.client.Next()
	if err != nil {
		return err
	}
	return c.GetStat()
}

// NextInstruction 是汇编层面的下一步，不进入函数
// 两种实现方法：
//  1. 判断指令是不是call，是的话就多执行一步 step-out
//  2. 在下一条指令下断点
//
// 当前使用：方法2
func (c *MyClient) NextInstruction() error {
	start := c.Current.Rip
	asms, err := c.Disassembly2(start, start+0x10)
	if err != nil {
		return err
	}
	var bPC uint64
	for i, asm := range asms {
		if asm.Loc.PC == c.Current.Rip {
			bPC = asms[i+1].Loc.PC
			err = c.CreateBreakpointByAddress(bPC, "")
			if err != nil {
				return err
			}
			break
		}
	}
	err = c.Continue()
	if err != nil {
		return err
	}
	return c.ClearBreakpointByAddress(bPC)
}

// Disassembly 是反汇编 Rip 寄存器附近的数据
func (c *MyClient) Disassembly() (api.AsmInstructions, error) {
	pc := c.Current.Rip
	asms, err := c.client.DisassemblePC(c.currentEvalScope(), pc, api.IntelFlavour)
	if err != nil {
		return nil, err
	}
	pcIndex := 0
	for index, asm := range asms {
		if asm.Loc.PC == pc {
			pcIndex = index
		}
	}
	if pcIndex < 4 {
		return asms, nil
	} else {
		return asms[pcIndex-4:], nil
	}
}

// Disassembly2 是反汇编 start 到 ends 范围内的数据
func (c *MyClient) Disassembly2(start, ends uint64) (api.AsmInstructions, error) {
	return c.client.DisassembleRange(c.currentEvalScope(), start, ends, api.IntelFlavour)
}

// StepInstruction 是汇编层面的单步运行
func (c *MyClient) StepInstruction() error {
	_, err := c.client.StepInstruction()
	if err != nil {
		return err
	}

	return c.GetStat()
}

// Step 是步入函数，会进入函数内部
func (c *MyClient) Step() error {
	_, err := c.client.Step()
	if err != nil {
		return err
	}
	return c.GetStat()
}

// StepOut 是跳出函数，会直接执行到调用者
func (c *MyClient) StepOut() error {
	_, err := c.client.StepOut()
	if err != nil {
		return err
	}
	return c.GetStat()
}

// ReadSourceCode 是读取源文件，并定位到行
func (c *MyClient) ReadSourceCode() ([]string, error) {
	return utils.ReadSourceCodeFromFile(c.Current.FilePath, c.Current.FileLine)
}

// ExamineMemory 用来读取特定地址的 n 个字节的数据
func (c *MyClient) ExamineMemory(address uint64, count int) ([]byte, error) {
	if count%0x10 != 0 {
		c := 0x10 - count%0x10
		count += c
	}
	memories, _, err := c.client.ExamineMemory(address, count)
	if err != nil {
		return nil, err
	}
	return memories, nil
}

// ClearBreakpointByName 是根据断点名消除断点
func (c *MyClient) ClearBreakpointByName(name string) error {
	var err error
	_, err = c.client.ClearBreakpointByName(name)
	return err
}

// ClearBreakpointByID 是根据 ID 消除断点
func (c *MyClient) ClearBreakpointByID(id int) error {
	_, err := c.client.ClearBreakpoint(id)
	return err
}

// ClearBreakpointByAddress 是根据地址删除断点
func (c *MyClient) ClearBreakpointByAddress(address uint64) error {
	points, err := c.ListBreakpoints()
	if err != nil {
		return err
	}
	for _, point := range points {
		if point.Addr == address {
			err = c.ClearBreakpointByID(point.ID)
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// ClearAllBreakpoints 删除所有 breakpoints
func (c *MyClient) ClearAllBreakpoints() error {
	points, err := c.ListBreakpoints()
	if err != nil {
		return err
	}
	for _, point := range points {
		if point.ID > 0 {
			err = c.ClearBreakpointByID(point.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *MyClient) ReRun(rebuild bool) error {
	_, err := c.client.Restart(rebuild)
	if err != nil {
		return err
	}
	err = c.Continue()
	if err != nil {
		return err
	}
	err = c.GetStat()
	if err != nil {
		return err
	}
	return err
}

func (c *MyClient) GetDataFromAddress(addr uint64, size int) (string, error) {
	data, err := c.ExamineMemory(addr, size)
	if err != nil {
		return "", err
	}
	builder := strings.Builder{}
	builder.WriteString("0x")
	for i := size - 1; i >= 0; i-- {
		_, err2 := fmt.Fprintf(&builder, "%x", data[i])
		if err2 != nil {
			return "", err2
		}
	}
	return builder.String(), nil
}
