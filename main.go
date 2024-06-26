package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/dweymouth/supersonic/backend"
	"github.com/dweymouth/supersonic/res"
	"github.com/dweymouth/supersonic/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// parse cmd line flags - see backend/cmdlineoptions.go
	flag.Parse()
	if *backend.FlagVersion {
		fmt.Println(res.AppVersion)
		return
	}
	if *backend.FlagHelp {
		flag.Usage()
		return
	}
	// rest of flag actions are handled in backend.StartupApp

	myApp, err := backend.StartupApp(res.AppName, res.DisplayName, res.AppVersionTag, res.LatestReleaseURL)
	if err != nil {
		if err != backend.ErrAnotherInstance {
			log.Fatalf("fatal startup error: %v", err.Error())
		}
		return
	}

	if myApp.Config.Application.UIScaleSize == "Smaller" {
		os.Setenv("FYNE_SCALE", "0.85")
	} else if myApp.Config.Application.UIScaleSize == "Larger" {
		os.Setenv("FYNE_SCALE", "1.1")
	}

	fyneApp := app.New()
	fyneApp.SetIcon(res.ResAppicon256Png)

	mainWindow := ui.NewMainWindow(fyneApp, res.AppName, res.DisplayName, res.AppVersion, myApp)
	myApp.OnReactivate = mainWindow.Show
	myApp.OnExit = mainWindow.Quit

	go func() {
		defaultServer := myApp.ServerManager.GetDefaultServer()
		if defaultServer == nil {
			mainWindow.Controller.PromptForFirstServer()
		} else {
			mainWindow.Controller.DoConnectToServerWorkflow(defaultServer)
		}

		// hacky workaround for https://github.com/fyne-io/fyne/issues/4964
		if runtime.GOOS == "linux" {
			time.Sleep(350 * time.Millisecond)
			canvas := mainWindow.Window.Canvas()
			size := canvas.Size()
			desired := mainWindow.DesiredSize()
			if !inDelta(size, desired, 1) {
				// window drawn at incorrect size on startup
				scale := canvas.Scale()
				SendResizeToPID(os.Getpid(), int(desired.Width*scale), int(desired.Height*scale))
			}
		}

	}()

	mainWindow.Show()
	mainWindow.Window.SetCloseIntercept(func() {
		mainWindow.SaveWindowSize()
		if myApp.Config.Application.CloseToSystemTray &&
			mainWindow.HaveSystemTray() {
			mainWindow.Window.Hide()
		} else {
			fyneApp.Quit()
		}
	})
	fyneApp.Run()

	log.Println("Running shutdown tasks...")
	myApp.Shutdown()
}

func inDelta(a, b fyne.Size, delta float32) bool {
	diffW := a.Width - b.Width
	diffH := a.Height - b.Height
	return diffW < delta && diffW > -delta &&
		diffH < delta && diffH > -delta
}
