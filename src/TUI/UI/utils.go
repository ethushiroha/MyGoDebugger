package UI

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/go-delve/delve/service/api"
	"strconv"
	"strings"
)

var wordList []string

func FormatASM(asms api.AsmInstructions, ip uint64) []string {
	result := make([]string, 0, 0)
	preFunc := ""
	funcLine := ""

	// 显示至多 17 行汇编代码
	lines := 17
	if len(asms) < lines {
		lines = len(asms)
	}

	for i := 0; i < lines; i++ {
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

func lastIndexOfNumberOrLetter(arg string) int {
	for i := 0; i < len(arg); i++ {
		if (arg[i] <= 'z' && arg[i] >= 'a') || (arg[i] >= '0' && arg[i] <= '9') {
			continue
		} else {
			return i
		}
	}
	return -1
}

func ParseArgs(arg string) (string, error) {
	unChangedString := strings.ToLower(arg)
	var result strings.Builder
	for len(unChangedString) > 0 {
		if unChangedString[0] == '$' {
			var regName string
			lastIndex := lastIndexOfNumberOrLetter(unChangedString[1:])
			if lastIndex == -1 {
				regName = unChangedString[1:]
				unChangedString = ""
			} else {
				regName = unChangedString[1 : 1+lastIndex]
				unChangedString = unChangedString[1+lastIndex:]
			}
			for _, reg := range client.Current.Regs {
				if strings.ToLower(reg.Name) == regName {
					tmp, err := strconv.ParseInt(reg.Value, 0, 64)
					if err != nil {
						return "", err
					}
					result.WriteString(fmt.Sprintf("%d", tmp))
					break
				}
			}
		} else if len(unChangedString) > 1 && unChangedString[0] == '0' && unChangedString[1] == 'x' {
			lastIndex := lastIndexOfNumberOrLetter(unChangedString[2:])
			var hex string
			if lastIndex == -1 {
				hex = unChangedString
				unChangedString = ""
			} else {
				hex = unChangedString[:lastIndex+2]
				unChangedString = unChangedString[lastIndex+2:]
			}
			value, err := strconv.ParseInt(hex, 0, 64)
			if err != nil {
				return "", err
			}
			result.WriteString(fmt.Sprintf("%d", value))
		} else {
			result.WriteByte(unChangedString[0])
			unChangedString = unChangedString[1:]
		}

	}
	return result.String(), nil
}

func CalculateAddress(expr string) (string, error) {
	e, err := ParseArgs(expr)
	if err != nil {
		return "", err
	}
	expression, err := govaluate.NewEvaluableExpression(e)
	if err != nil {
		return "", err
	}
	result, err := expression.Evaluate(nil)
	if err != nil {
		return "", err
	}
	address := fmt.Sprintf("0x%x", int(result.(float64)))
	return address, nil
}
