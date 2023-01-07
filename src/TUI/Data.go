package main

import MyApi "MyDebugger/src/api"

type CurrentData struct {
	Client          *MyApi.MyClient
	DisassemblyData []string
	RegsData        []string
	MemoryData      []string
	StackData       []string
}

func (data *CurrentData) Disassembly() error {
	asms, err := data.Client.Disassembly()
	if err != nil {
		return err
	}
	points, err := data.Client.ListBreakpoints()
	if err != nil {
		return err
	}
	data.DisassemblyData = ASMToStrings(asms, points, data.Client.Current.Rip)
	return nil
}

func (data *CurrentData) Registers() {
	data.RegsData = RegsToStrings(data.Client.Current.Regs)
}

func (data *CurrentData) ExamineMemory(start, mode uint64, format string) error {
	ends := start + 0x80
	mems, err := data.Client.ExamineMemory(start, int(ends-start))
	if err != nil {
		return err
	}
	data.MemoryData = FormatMemory(mems, start, mode, format)
	return nil
}

func (data *CurrentData) ExamineStack(mode uint64, format string) error {
	start := data.Client.Current.Rsp
	return data.ExamineMemory(start, mode, format)
}

func (data *CurrentData) FlashData() error {
	err := data.Disassembly()
	if err != nil {
		return err
	}
	data.Registers()
	err = data.ExamineStack(8, "hex")
	if err != nil {
		return err
	}
	return nil
}

func InitData(address string) (*CurrentData, error) {
	data := new(CurrentData)
	var err error
	data.Client, err = MyApi.NewClientWithMain(address)
	if err != nil {
		return nil, err
	}
	err = data.FlashData()
	if err != nil {
		return nil, err
	}
	return data, nil
}
