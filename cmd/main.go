package main

import (
	"fmt"
	"os"

	"github.com/MyonKeminta/sysbench-analyzer/lib"
)

func main() {
	// TODO: Process command line args more elegantly
	// TODO: --help
	var command string
	if len(os.Args) <= 1 {
		command = "check"
	} else {
		command = os.Args[1]
	}

	args := os.Args[2:]

	switch command {
	case "check":
		check(args)
		break
	case "plot":
		plot(args)
		break
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %v\nUsage: %v {check|plot} [args...]\n", command, os.Args[0])
	}

}

func check(args []string) {
	fmt.Printf("Check from stdin for QPS dropping, threshold 70%%, ignoring invalid lines...\n")

	success, violations := lib.CheckQPSDropFromSysbenchOutputStream(os.Stdin, 0.7)

	if success {
		fmt.Println("Passed.")
	} else {
		fmt.Println("Abnormal Lines:")
		fmt.Print(violations)
	}
}

func plot(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Missing file name.\nUsage: %v plot <output-img-file>", os.Args[0])
	}
	imageFile := args[0]

	analyzer := lib.NewSysbenchAnalyzer([]lib.Rule{}, true)
	records, _ := analyzer.ParseToEnd(os.Stdin)

	err := lib.PlotQPS(records, imageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Plotting failed. Error:\n%v\n", err)
	}
}
