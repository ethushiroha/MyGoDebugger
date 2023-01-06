package main

import (
	"fmt"
	"github.com/go-delve/delve/service/api"
	"strings"
)

func ASMToStrings(asms api.AsmInstructions, pointers []*api.Breakpoint, ip uint64) []string {
	result := make([]string, 0)
	for _, asm := range asms {
		pc := asm.Loc.PC

		line := fmt.Sprintf("0x%x    [p]    %s", pc, asm.Text)
		for _, pointer := range pointers {
			pointerPC := pointer.Addr
			if pointerPC == pc {
				line = strings.Replace(line, "[p]", "#", 1)
				break
			}
		}
		line = strings.Replace(line, "[p]", " ", 1)
		if pc == ip {
			line = "[red]" + line + "[white]"
		}
		result = append(result, line)
	}
	return result
}

func RegsToStrings(regs api.Registers) []string {
	result := make([]string, 0)
	for i := 0; i <= 0x10; i++ {
		line := fmt.Sprintf("%-3s     %s", regs[i].Name, regs[i].Value)
		result = append(result, line)
	}
	return result
}

func MemoryToStrings(mems []byte, start uint64) []string {
	result := make([]string, 0)
	var offset uint64
	for len(mems) > 0 {
		data1 := fmt.Sprintf("0x%02x%02x%02x%02x%02x%02x%02x%02x", mems[7], mems[6], mems[5], mems[4], mems[3], mems[2], mems[1], mems[0])
		data2 := fmt.Sprintf("0x%02x%02x%02x%02x%02x%02x%02x%02x", mems[15], mems[14], mems[13], mems[12], mems[11], mems[10], mems[9], mems[8])
		line := fmt.Sprintf("0x%x    %s  %s", start+offset, data1, data2)
		result = append(result, line)
		mems = mems[0x10:]
		offset += 0x10
	}
	return result
}
