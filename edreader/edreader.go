package edreader

import (
	"bufio"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-cmp/cmp"
	"github.com/peterbn/EDx52display/mfd"

	"github.com/peterbn/EDx52display/conf"
)

const DisplayPages = 3

var tick time.Ticker

const (
	pageTargetInfo = iota
	pageLocation
	pageCargo
)

// Mfd is the MFD display structure that will be used by this module. The number of pages should not be changed
var Mfd = mfd.Display{Pages: make([]mfd.Page, DisplayPages)}

// MfdLock locks the current MFD for reads and writes
var MfdLock = sync.RWMutex{}

// PrevMfd is the previous Mfd written to file, to be used for comparisons and avoid superflous updates.
var PrevMfd = Mfd.Copy()

var (
	lastJournalFile   string
	lastJournalOffset int64
	lastJournalInode  uint64       // for file rotation detection (optional, platform-specific)
	lastJournalState  Journalstate // <-- Add this to persist state
)

// Start starts the Elite Dangerous journal reader routine
func Start(cfg conf.Conf) {
	// Update immediately, to ensure the mfd.json file exist
	log.Info("Starting journal listener")
	journalfolder := cfg.ExpandJournalFolderPath()
	log.Debugln("Looking for journal files in " + journalfolder)
	updateMFD(journalfolder)
	tick := time.NewTicker(time.Duration(cfg.RefreshRateMS) * time.Millisecond)

	go func() {
		for range tick.C {
			updateMFD(journalfolder)
		}
	}()
}

func updateMFD(journalfolder string) {
	journalFile := findJournalFile(journalfolder)
	handleJournalFileIncremental(journalFile)

	handleModulesInfoFile(filepath.Join(journalfolder, FileModulesInfo))
	handleCargoFile(filepath.Join(journalfolder, FileCargo))
	swapMfd()
}

// handleJournalFileIncremental reads only new lines from the journal file since the last read.
func handleJournalFileIncremental(filename string) {
	if filename == "" {
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Warnln("Error opening journal file ", filename, err)
		return
	}
	defer file.Close()

	var offset int64 = 0
	if filename == lastJournalFile {
		offset = lastJournalOffset
	}

	info, err := file.Stat()
	if err != nil {
		log.Warnln("Error stating journal file ", filename, err)
		return
	}

	// If file shrank (rotated), start from beginning
	if offset > info.Size() {
		offset = 0
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		log.Warnln("Error seeking journal file ", filename, err)
		return
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	state := lastJournalState // Start from last known state
	linesRead := 0
	for scanner.Scan() {
		ParseJournalLine(scanner.Bytes(), &state)
		linesRead++
	}
	if linesRead > 0 {
		lastJournalState = state // Only update if new lines were read
	}
	RefreshDisplay(lastJournalState)

	// Save offset for next time
	pos, _ := file.Seek(0, 1)
	lastJournalFile = filename
	lastJournalOffset = pos
}

// Stop closes the watcher again
func Stop() {
	tick.Stop()
}

func findJournalFile(folder string) string {
	// Based on https://github.com/EDCD/EDMarketConnector/blob/693463d3a0dbe58a1f72b83fc09a44a4398af3fd/monitor.py#L264
	// because I don't have Odyssey myself
	files, _ := filepath.Glob(filepath.Join(folder, "Journal.*.*.log"))

	var lastFileTime time.Time
	var mostRecentJournal = ""

	for _, filename := range files {
		info, err := os.Stat(filename)
		if err != nil {
			continue
		}
		if mostRecentJournal == "" || info.ModTime().After(lastFileTime) {
			lastFileTime = info.ModTime()
			mostRecentJournal = filename
		}
	}
	return mostRecentJournal
}

func swapMfd() {
	MfdLock.RLock()
	defer MfdLock.RUnlock()
	eq := cmp.Equal(Mfd, PrevMfd)
	if !eq {
		mfd.Write(Mfd)
		PrevMfd = Mfd.Copy()
	}
}
