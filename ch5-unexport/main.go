package main

import (
	"fmt"
	"os"
	"github.com/mhrabovcin/golang-challenge/ch5-unexport/unexport"
)

func main() {

	config := unexport.NewConfig("github.com/mhrabovcin/golang-challenge/ch5-unexport/scanned/one");
	config.Workspace = false
	unexporter, err := unexport.NewUnexporter(config)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("unused ", len(unexporter.GetUnusedNames()), " of ", len(unexporter.UsageStats.Identifiers), " identifiers")
	fmt.Println("\nrename commands: \n")

	for _, command := range unexporter.GenerateRenameCommands() {
		fmt.Println(command)
	}

}
