package Entities

import (
	"fmt"
	"main/Shared"
	"strconv"
	"time"
)

type InputOutput struct {
	Input  int
	Output float64
}

type SystemModeling struct {
	sampling      int
	inputName     string
	outputName    string
	csvWriter     Shared.CsvHandler
	minInputValue float64
	maxInputValue float64
}

func CreateSystemModeling(samplingRate int, inputMetric string, outputMetric string, outputFile string, minInputValue float64, maxInputValue float64) SystemModeling {

	obj := SystemModeling{maxInputValue: maxInputValue, minInputValue: minInputValue, inputName: inputMetric, outputName: outputFile, sampling: samplingRate}
	obj.csvWriter = Shared.CsvHandler{}
	obj.csvWriter.CreateFile(outputFile, []string{"timestamp", "serviceName", inputMetric, outputMetric})

	return obj

}

func (s *SystemModeling) GetMinInputValue() float64 {
	return s.minInputValue
}

func (s *SystemModeling) GetMaxInputValue() float64 {
	return s.maxInputValue
}

func (s *SystemModeling) writeInputOutputValue(input float64, serviceName string, output float64) {
	now := time.Now()

	inputStr := strconv.FormatFloat(float64(input), 'f', 2, 64)
	outputStr := strconv.FormatFloat(float64(output), 'f', 4, 64)

	s.csvWriter.WriteLine([]string{now.Format("2006-01-02 15:04:05"), serviceName, inputStr, outputStr})
}

func (s *SystemModeling) Capture(data map[string]InputOutput) {
	for key, value := range data {
		fmt.Println("Writing", "Input", float64(value.Input), "Output", value.Output/float64(s.sampling), "service", key)
		s.writeInputOutputValue(float64(value.Input), key, value.Output/float64(s.sampling))

	}
}

func (s *SystemModeling) FinalizeExperiment() {
	s.csvWriter.CloseFile()
}

/*func main() {

	test := CreateSystemModeling(10, "CPU", "RT", "test-int.csv", 256, 1000)
	data := make(map[string]InputOutput)
	data["test"] = InputOutput{Input: 10, Output: 20}
	test.Capture(data)

}*/
