package UI

import (
	"MyDebugger/src/utils"
	"fmt"
)

type Tracker struct {
	data      string
	size      int
	isChanged bool
}

type Trackers map[string]*Tracker

var trackers = NewTrackers()

func NewTrackers() Trackers {
	return make(Trackers)
}

func (t *Trackers) add(address string, size int) {
	if _, ok := (*t)[address]; ok {
		return
	}
	u64Address, err := utils.StringToUint64(address)
	if err != nil {
		return
	}
	d, err := client.GetDataFromAddress(u64Address, size)
	if err != nil {
		return
	}
	(*t)[address] = &Tracker{
		size:      size,
		data:      d,
		isChanged: false,
	}
}

func (t *Trackers) getTrackersData() []string {
	result := make([]string, 0)
	for addr, tracker := range *t {
		line := fmt.Sprintf("%s  %s", addr, tracker.data)
		if tracker.isChanged {
			line = fmt.Sprintf("[red]%s[white]", line)
			tracker.isChanged = false
		}
		result = append(result, line)
	}
	return result
}

func (t *Trackers) track() bool {
	flag := false
	for addr, tracker := range trackers {
		d, err := client.GetDataFromStringAddress(addr, tracker.size)
		if err != nil {
			return flag
		}
		if d != tracker.data {
			tracker.isChanged = true
			tracker.data = d
			flag = true
		}
	}
	return flag
}

func (t *Trackers) remove(address string) {
	delete(*t, address)
}
