package utils

import (
	"fmt"
	"log"
	"os"
)

func SetLogger(path string) {
	removeLogFile(path)
	file := createLogFile(path)
	log.SetOutput(file)
}

func removeLogFile(path string) {
	err := os.Remove(path)

	if err != nil {
		fmt.Println(err)
		return
	}
}

func createLogFile(path string) *os.File {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}