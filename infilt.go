package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/busy"
	"github.com/claude42/infiltrator/ui"
	// dateparser "github.com/markusmobius/go-dateparser"
)

func main() {
	err := run()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.User()
	err := config.Load()
	if err != nil {
		return err
	}
	defer config.WriteStateFile()

	// debug log

	if !cfg.Debug {
		log.SetOutput(io.Discard)
	} else {
		debug, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		fail.OnError(err, "Failed to open debug log file")
		defer debug.Close()
		log.SetOutput(debug)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	config.PostEventFunc = ui.InfiltPostEvent

	ctx, cancelFunc := context.WithCancel((context.Background()))
	var wg sync.WaitGroup

	// Busy spinner first :-)
	wg.Add(1)
	go busy.StartBusySpinner(ctx, &wg)

	quit := make(chan string, 10)
	fm := model.NewFilterManager(ctx, &wg, quit)

	// Set up UI
	window := ui.Setup()

	wg.Add(1)
	go window.MetaEventLoop(ctx, &wg, quit)

	if cfg.Stdin {
		fm.ReadFromStdin()
	} else {
		fm.ReadFromFile(cfg.FilePath)
	}

	wg.Add(1)
	go fm.EventLoop()

	window.CreatePresetPanels()

	// wait for UI thread to finish

	var message string
	for message = range quit {
		log.Printf("Quit message: %s", message)
	}

	cancelFunc()
	wg.Wait()

	ui.Cleanup()

	fmt.Fprintln(os.Stderr, message)

	return nil
}
