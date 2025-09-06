package Shared

//package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

type CsvHandler struct {
	filename string
	header   []string
	file     *csv.Writer
	arquivo  *os.File
}

func (c *CsvHandler) CloseFile() {
	defer c.file.Flush()
	defer c.arquivo.Close()
}

func (c *CsvHandler) CreateFile(filename string, header []string) {
	c.filename = filename
	c.header = header
	arquivo, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error to create file:", err)
	}
	// Cria um escritor CSV
	c.file = csv.NewWriter(arquivo)

	err = c.file.Write(c.header)
	if err != nil {
		log.Fatal("Error to write header in file", err)
	}
	fmt.Println("arquivo criado", c.header)
	defer c.file.Flush()

}

func (c *CsvHandler) WriteLine(line []string) {

	err := c.file.Write(line)
	if err != nil {
		log.Fatal("Error to write line in file:", err)
	}
	defer c.file.Flush()
}

func main() {

	csvA := CsvHandler{}
	csvA.CreateFile("testefile.csv", []string{"aaa", "bbb"})
	csvA.WriteLine([]string{"asdasda", "dasdadad"})
	csvA.CloseFile()
}
