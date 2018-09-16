package main

import (
	"fmt"
	"os"

	"github.com/MyonKeminta/sysbench-analyzer/lib"
)

func main() {
	fmt.Printf("Check from stdin for QPS dropping, threshold 70%%, ignoring invalid lines...\n")

	success, violations := lib.CheckQPSDropFromSysbenchOutputStream(os.Stdin, 0.7)

	if success {
		fmt.Println("Passed.")
	} else {
		fmt.Println("Abnormal Lines:")
		fmt.Print(violations)
	}
}
