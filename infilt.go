package main

import (
	// "flag"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	//"time"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/ui"

	flag "github.com/spf13/pflag"
)

func main() {
	err := run()

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
	flag.BoolVarP(&cm.Debug, "debug", "d", false, "Log debugging information to ./debug.log")

	flag.Parse()

	// debug log

	if !config.GetConfiguration().Debug {
		log.SetOutput(io.Discard)
	} else {
		debug, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Panicf("Failed to open debug log file: %v", err)
		}
		defer debug.Close()
		log.SetOutput(debug)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	cfg := config.GetConfiguration()

	cfg.Context, cfg.Cancel = context.WithCancel((context.Background()))

	// Set up UI
	window := ui.Setup()

	cfg.Quit = make(chan string, 10)
	cfg.WaitGroup.Add(1)
	go window.MetaEventLoop()

	fm := model.GetFilterManager()
	cfg.PostEventFunc = ui.InfiltPostEvent

	switch len(flag.Args()) {
	case 0:
		cm.FileName = "[stdin]"
		cm.FilePath = ""
		cm.Stdin = true
		fm.ReadFromStdin()
	case 1:
		cm.FilePath = flag.Args()[0]
		cm.FileName = filepath.Base(cm.FilePath)
		cm.Stdin = false

		fm.ReadFromFile(cm.FilePath)
	default:
		flag.Usage()
		return fmt.Errorf("try again")
	}

	cfg.WaitGroup.Add(1)
	go fm.EventLoop()

	// wait for UI thread to finish

	var message string
	for message = range cfg.Quit {
		log.Printf("in loop %s", message)
	}

	cfg.Cancel()
	cfg.WaitGroup.Wait()

	ui.Cleanup()

	fmt.Fprintln(os.Stderr, message)

	return nil

	// Set up filtering pipeline
	// pipeline := model.GetPipeline()

	// contentUpdate := make(chan []model.Line, 10)

	// go model.GetFilterManager().EventLoop(contentUpdate)
	// filePath := flag.Args()[0]
	// go model.ReadFromFile(filePath, contentUpdate)

	// for i := 0; i < 3; i++ {
	// 	time.Sleep(time.Second)
	// 	log.Println("waiting")
	// }

	// switch len(flag.Args()) {
	// case 0:
	// 	cm.FileName = "[stdin]"
	// 	go buffer.ReadFromStdin(ui.GetScreen().PostEvent)
	// case 1:
	// 	filePath := flag.Args()[0]
	// 	cm.FileName = filepath.Base(filePath)
	// 	go buffer.ReadFromFile(filePath, ui.GetScreen().PostEvent)
	// default:
	// 	flag.Usage()
	// 	return fmt.Errorf("Try again")
	// }

	// // Set up UI
	// window := ui.Setup()
	// defer ui.Cleanup()

	// quit := make(chan struct{})
	// go window.EventLoop(quit)

	// // wait for UI thread to finish
	// <-quit

	// return nil
}
