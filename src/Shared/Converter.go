package Shared

import "strconv"

func ConvertStringToInt(input string) int {
	num, err := strconv.Atoi(input)
	if err != nil {
		// Se ocorrer um erro na conversão
		panic("Error to convert string to number:" + err.Error())

	}
	return num
}
