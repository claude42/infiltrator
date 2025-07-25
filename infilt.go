package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	//"time"

	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/ui"
	//"github.com/claude42/infiltrator/util"
)

// Command line options
var showLineNumbers = false

func main() {
	var err error
	// Set up logging first
	debug, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Panicf("Failed to open debug log file: %v", err)
	}
	defer debug.Close()
	log.SetOutput(debug)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Started")

	err = run()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {

	// Parse command line
	flag.BoolVar(&showLineNumbers, "lines", false, "Show line numbers")

	flag.Parse()
	if len(flag.Args()) != 1 {
		return fmt.Errorf("usage: %s filename", filepath.Base(os.Args[0]))
	}
	filePath := flag.Args()[0]

	// Set up filtering pipeline
	p := model.GetPipeline()

	// Create buffer
	buffer, err := model.NewBufferFromFile(filePath)
	if err != nil {
		return err
	}
	p.AddFilter(buffer)

	// Set up UI
	window := ui.Setup(p)
	defer ui.Cleanup()
	window.ShowLineNumbers(showLineNumbers)

	quit := make(chan struct{})
	go window.EventLoop(quit)

	<-quit

	return nil

	// for {
	// 	select {
	// 	case <-quit:
	// 		return nil
	// 	}
	// }
}
