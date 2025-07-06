package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	_ "embed"

	"github.com/getlantern/systray"
	"github.com/pellux-network/EDx52display/conf"
	"github.com/pellux-network/EDx52display/edreader"
	"github.com/pellux-network/EDx52display/edsm"
	"github.com/pellux-network/EDx52display/mfd"
)

// TextLogFormatter gives me custom command-line formatting
type TextLogFormatter struct{}

//go:embed icon.ico
var iconData []byte

func (f *TextLogFormatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := time.Now().UTC().Format("2006-01-02 15:04:05")
	level := entry.Level.String()
	message := entry.Message

	return []byte(timestamp + " - " + strings.ToUpper(level) + " - " + message + "\n"), nil
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// Set up systray icon and menu
	systray.SetIcon(getIcon()) // You can provide your own icon as []byte
	systray.SetTitle("EDx52Display")
	systray.SetTooltip("EDx52Display is running")

	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	// Start your main logic in a goroutine
	go func() {
		defer func() {
			// Attempt to catch any crash messages before the cmd window closes
			if r := recover(); r != nil {
				log.Warnln("Crashed with message")
				log.Warnln(r)
			}
		}()
		var logLevelArg string
		flag.StringVar(&logLevelArg, "log", "trace", "Desired log level. One of [panic, fatal, error, warning, info, debug, trace]. Default: trace.")

		flag.Parse()
		logLevel, err := log.ParseLevel(logLevelArg)
		if err != nil {
			log.Panic(err)
		}

		log.SetLevel(logLevel)
		log.SetFormatter(&TextLogFormatter{})

		// Ensure logs directory exists
		logDir := "logs"
		_ = os.MkdirAll(logDir, 0755)
		logFileName := time.Now().Format("2006-01-02_15.04.05") + ".log"
		logPath := filepath.Join(logDir, logFileName)

		// Set up log rotation
		log.SetOutput(&lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    10, // megabytes
			MaxBackups: 5,
			MaxAge:     30,   //days
			Compress:   true, // compress old logs
		})

		log.Infof("Logging to %s", logPath)

		conf := conf.LoadConf()

		// Calculate number of enabled pages
		pageCount := 0
		for _, enabled := range conf.Pages {
			if enabled {
				pageCount++
			}
		}

		err = mfd.InitDevice(uint32(pageCount), edsm.ClearCache)
		if err != nil {
			log.Panic(err)
		}
		defer mfd.DeInitDevice()

		edreader.Start(conf)
		defer edreader.Stop()

		// Wait for quit
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	// Optionally, handle other menu items here
}

func onExit() {
	// Cleanup tasks if needed
}

// getIcon returns an icon as []byte. Replace with your own icon if desired.
func getIcon() []byte {
	return iconData
}
