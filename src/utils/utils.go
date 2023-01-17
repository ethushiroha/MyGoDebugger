package utils

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"unsafe"
)

func PrintArrayWithDetail(array ...any) {
	for _, i := range array {
		fmt.Printf("%+v\n", i)
	}
}

func PrintError(err error) {
	fmt.Printf("%+v\n", err)
}

func PrintArray(array ...any) {
	for _, i := range array {
		fmt.Println(i)
	}
}

func ReadSourceCodeFromFile(filename string, line int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	result := make([]string, 0)
	start := "+-----------------------------------------------------------+"
	title := "|  line   |"
	result = append(result, start)
	result = append(result, title)
	result = append(result, start)
	scanner := bufio.NewScanner(file)
	for i := 1; i <= line-5; i++ {
		scanner.Scan()
	}
	// todo: 自动找出行号长度，进行对齐
	for i := line - 4; i < line+11; i++ {
		if scanner.Scan() {
			var text string
			if i == line {
				text = fmt.Sprintf("| >%-5d  |", i)
			} else {
				text = fmt.Sprintf("|  %-5d  |", i)
			}
			text += scanner.Text()
			result = append(result, text)
		}
	}
	result = append(result, start)
	return result, nil

}

func StringToUint64(data string) (uint64, error) {
	// base 表示进制，2-64，如果为0，会自己判断，0x 为 16进制，0 为 8进制，否则为 10进制
	return strconv.ParseUint(data, 0, 64)
}

func GetStructPtrUnExportedField(source any, fieldName string) reflect.Value {
	// 获取非导出字段反射对象
	v := reflect.ValueOf(source).Elem().FieldByName(fieldName)
	// 构建指向该字段的可寻址（addressable）反射对象
	return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
}

func TripHexZero(hex string) (string, error) {
	tmp, err := StringToUint64(hex)
	if err != nil {
		return "", err
	}
	result := fmt.Sprintf("0x%x", tmp)
	return result, nil
}
