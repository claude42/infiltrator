package main

import (
	// "flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	//"time"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/ui"

	//"github.com/claude42/infiltrator/util"

	flag "github.com/spf13/pflag"
)

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
	cm := config.GetConfiguration()

	// Parse command line
	flag.BoolVarP(&cm.ShowLineNumbers, "lines", "l", false, "Show line numbers")
	flag.BoolVarP(&cm.FollowFile, "follow", "f", false, "Follow changes to file")

	flag.Parse()
	// if len(flag.Args()) != 1 {
	// 	return fmt.Errorf("usage: %s filename", filepath.Base(os.Args[0]))
	// }

	// Set up filtering pipeline
	pipeline := model.GetPipeline()

	// Create buffer + load file
	buffer := model.NewBuffer()
	pipeline.AddFilter(buffer)

	switch len(flag.Args()) {
	case 0:
		cm.FileName = "[stdin]"
		go buffer.ReadFromStdin(ui.GetScreen().PostEvent)
	case 1:
		filePath := flag.Args()[0]
		cm.FileName = filepath.Base(filePath)
		go buffer.ReadFromFile(filePath, ui.GetScreen().PostEvent)
	default:
		return fmt.Errorf("usage: %s [-f] [-l] [filename]", filepath.Base(os.Args[0]))
	}

	// Set up UI
	window := ui.Setup()
	defer ui.Cleanup()

	quit := make(chan struct{})
	go window.EventLoop(quit)

	// wait for UI thread to finish
	<-quit

	return nil
}
