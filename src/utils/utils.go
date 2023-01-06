package utils

import (
	"bufio"
	"fmt"
	"os"
)

func PrintArrayWithDetail[T interface{}](array []T) {
	for _, i := range array {
		fmt.Printf("%+v\n", i)
	}
}

func PrintError(err error) {
	fmt.Printf("%+v\n", err)
}

func PrintArray[T any](array []T) {
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
