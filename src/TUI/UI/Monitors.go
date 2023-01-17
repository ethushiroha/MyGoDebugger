package UI

import (
	"MyDebugger/src/utils"
	"fmt"
)

type Monitor struct {
	size         int
	data         string
	isBreakpoint bool
	isChanged    bool
}

type Monitors map[string]*Monitor

var monitors = NewMonitors()

func NewMonitors() Monitors {
	return make(map[string]*Monitor)
}

func (m *Monitors) getMonitorsData() []string {
	result := make([]string, 0)
	for address, monitor := range monitors {
		line := fmt.Sprintf("%s  %s", address, monitor.data)
		if monitor.isChanged {
			line = fmt.Sprintf("[red]%s[white]", line)
			monitor.isChanged = false
		}
		result = append(result, line)
	}
	return result
}

func (m *Monitors) monitorAddress() bool {
	flag := false
	for address, monitor := range monitors {
		u64Address, err := utils.StringToUint64(address)
		if err != nil {
			return false
		}
		data, err := client.GetDataFromAddress(u64Address, monitor.size)
		if err != nil {
			return false
		}
		if monitor.data != "" && monitor.data != data {
			monitor.isChanged = true
			flag = true
		}
		monitor.data = data
	}
	return flag
}

func (m *Monitors) add(name string, size int) {
	(*m)[name] = &Monitor{
		size:         size,
		data:         "",
		isBreakpoint: false,
		isChanged:    false,
	}
}
