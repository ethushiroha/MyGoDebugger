package main

import (
	"fmt"
	"github.com/go-delve/delve/service/api"
	"strings"
)

var wordList []string

func FormatASM(asms api.AsmInstructions, ip uint64) []string {
	result := make([]string, 0, 0)
	preFunc := ""
	funcLine := ""

	for i := 0; i < len(asms); i++ {
		line := ""
		functionName := asms[i].Loc.Function.Name()
		if functionName != preFunc {
			if functionName != "" {
				funcLine = fmt.Sprintf("[yellow]; Function %s [white]", functionName)
			} else {
				funcLine = "[yellow];  [white]"
			}
			result = append(result, funcLine)
			preFunc = functionName
		}
		pc := asms[i].Loc.PC
		line = fmt.Sprintf("0x%x    [p]    %s", pc, asms[i].Text)
		if asms[i].Breakpoint {
			line = strings.Replace(line, "[p]", "#", 1)
		} else {
			line = strings.Replace(line, "[p]", " ", 1)
		}

		if pc == ip {
			line = "[red]" + line + "[white]"
		}
		result = append(result, line)
	}
	return result
}

func RegsToStrings(regs api.Registers) []string {
	result := make([]string, 0, 0)
	for i := 0; i <= 16; i++ {
		line := fmt.Sprintf("%-3s     %s", regs[i].Name, regs[i].Value)
		result = append(result, line)
	}
	return result
}

func StacktraceToStrings(stacktrace []api.Stackframe) []string {
	result := make([]string, 0, 0)
	for _, stack := range stacktrace {
		line := fmt.Sprintf("%s:%d", stack.Function.Name(), stack.Line)
		result = append(result, line)
	}

	return result
}

func BreakpointsToStrings(breakpoints []*api.Breakpoint) []string {
	result := make([]string, 0, 0)
	for _, point := range breakpoints {
		if point.ID < 0 {
			continue
		}
		line := fmt.Sprintf("%02d | %s:%d", point.ID, point.FunctionName, point.Line)
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
	result := make([]string, 0, 0)
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

func StringsToString(data []string) string {
	//return data[0]
	result := strings.Builder{}
	for _, d := range data {
		result.WriteString(d)
		result.WriteByte('\n')
	}
	return result.String()
}

// getDicKeys 从字典里获取所有的 key， 这里用来生成命令提示信息
func getDicKeys[T any](dic map[string]T) []string {
	keys := make([]string, len(dic))
	i := 0
	for k := range dic {
		keys[i] = k
		i++
	}
	return keys
}

// AutoComplete 命令提示
func AutoComplete(current string) []string {
	result := make([]string, 0, 0)
	if len(current) == 0 {
		return nil
	}
	for _, word := range wordList {
		if strings.HasPrefix(strings.ToLower(word), strings.ToLower(current)) {
			result = append(result, word)
		}
	}
	if len(result) < 1 {
		return nil
	}
	return result
}
