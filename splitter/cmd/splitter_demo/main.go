package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/stevegt/splitter"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) != 2 {
		fmt.Println("Invalid arguments. Please provide a file path and chunk number.")
		os.Exit(1)
	}

	filePath := args[0]
	chunkNum, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("Invalid chunk number. Please provide a valid number.")
		os.Exit(1)
	}

	fs := splitter.NewFileSplitter(filePath)
	chunks, err := fs.SplitFile()
	if err != nil {
		fmt.Printf("Error parsing file: %v\n", err)
		os.Exit(1)
	}

	if chunkNum <= 0 || chunkNum > len(chunks) {
		fmt.Printf("Chunk number out of range. Please provide a number between 1 and %d.\n", len(chunks))
		os.Exit(1)
	}

	fmt.Println(chunks[chunkNum-1])
}
