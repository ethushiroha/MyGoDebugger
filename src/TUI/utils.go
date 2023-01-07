package main

import (
	"fmt"
	"github.com/go-delve/delve/service/api"
	"strconv"
	"strings"
)

func stringToUint64(data string) (uint64, error) {
	return strconv.ParseUint(data, 0, 64)
}

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

func CountInLineWithMode(mode uint64) uint64 {
	switch mode {
	case 1:
		return 8
	case 2:
		return 4
	case 4:
		return 4
	case 8:
		return 2
	default:
		return 0
	}
}

func getFormat(signal byte) string {
	format := ""
	switch signal {
	case 'x':
		format = "hex"
	case 'd':
		format = "dec"
	}
	return format
}

func getMode(signal byte) uint64 {
	var mode uint64 = 0
	switch signal {
	case 'g':
		mode = 8
	case 'w':
		mode = 4
	case 'h':
		mode = 2
	case 'b':
		mode = 1
	}
	return mode
}

// FormatMemory 格式化数据，例如 x gx addr 会变成每行 2 个 8 字节的数据
func FormatMemory(mems []byte, start uint64, mode uint64, format string) []string {
	result := make([]string, 0)
	var offset uint64
	for len(mems) > 0 {
		sb := strings.Builder{}
		address := fmt.Sprintf("0x%x", start+offset)
		sb.WriteString(address)
		sb.WriteString("    ")
		count := CountInLineWithMode(mode)
		var j uint64
		for j = 0; j < count; j++ {
			sb.WriteString("0x")
			var i uint64
			for i = 0; i < mode; i++ {
				index := (j+1)*mode - i - 1
				if format == "hex" {
					sb.WriteString(fmt.Sprintf("%02x", mems[index]))
				}
				// todo: make dec format
				//else if format == "dec" {
				//	sb.WriteString(fmt.Sprintf("%03d", mems[index]))
				//}
			}
			sb.WriteString(" ")
		}
		mems = mems[count*mode:]
		offset += count * mode
		result = append(result, sb.String())
	}
	return result
}
