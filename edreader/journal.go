package edreader

import (
	"bufio"
	"io"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/buger/jsonparser"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// LocationType indicates where in a system the player is
type LocationType int

const (
	// LocationSystem means the player is somewhere in the system, not close to a body
	LocationSystem LocationType = iota
	// LocationPlanet means the player is close to a planetary body
	LocationPlanet
	// LocationLanded indicates the player has touched down
	LocationLanded
	// LocationDocked indicates the player has docked at a station (or outpost)
	LocationDocked
)

// Journalstate encapsulates the player state baed on the journal
type Journalstate struct {
	Location
	EDSMTarget
	Destination // NEW: add destination info
}

// Location indicates the players current location in the game
type Location struct {
	Type LocationType

	SystemAddress int64
	StarSystem    string

	Body     string
	BodyID   int64
	BodyType string

	Latitude  float64
	Longitude float64
}

// EDSMTarget indicates a system targeted by the FSD drive for a jump
type EDSMTarget struct {
	Name          string
	SystemAddress int64
}

// Destination holds the current destination info from Status.json
type Destination struct {
	SystemID int64
	BodyID   int64
	Name     string
}

const (
	systemaddress = "SystemAddress"
	bodyid        = "BodyID"
	starsystem    = "StarSystem"
	docked        = "Docked"
	body          = "Body"
	bodytype      = "BodyType"
	bodyname      = "BodyName"
	stationname   = "StationName"
	stationtype   = "StationType"
	latitude      = "Latitude"
	longitude     = "Longitude"
	name          = "Name"
)

type parser struct {
	line []byte
}

func (p *parser) getString(field string) (string, bool) {
	str, err := jsonparser.GetString(p.line, field)
	if err != nil {
		return "", false
	}
	return str, true
}

func (p *parser) getInt(field string) (int64, bool) {
	num, err := jsonparser.GetInt(p.line, field)
	if err != nil {
		return 0, false
	}
	return num, true
}

func (p *parser) getBool(field string) (bool, bool) {
	b, err := jsonparser.GetBoolean(p.line, field)
	if err != nil {
		return false, false
	}
	return b, true
}

func (p *parser) getFloat(field string) (float64, bool) {
	f, err := jsonparser.GetFloat(p.line, field)
	if err != nil {
		return 0, false
	}
	return f, true
}

var printer = message.NewPrinter(language.English)

var (
	lastJournalFile    string
	lastJournalOffset  int64
	lastJournalState   Journalstate
	lastStatusFileSize int64 // NEW: for status file change detection
)

// handleJournalFile reads only new lines from the journal file since the last read.
func handleJournalFile(filename string) {
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

	// Save offset for next time
	pos, _ := file.Seek(0, 1)
	lastJournalFile = filename
	lastJournalOffset = pos
}

// handleStatusFile reads Status.json for the current destination
func handleStatusFile(filename string) {
	if filename == "" {
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return
	}
	if info.Size() == lastStatusFileSize {
		return
	}
	lastStatusFileSize = info.Size()

	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	destObj, _, _, err := jsonparser.Get(data, "Destination")
	if err == nil && len(destObj) > 0 {
		sysID, _ := jsonparser.GetInt(destObj, "System")
		bodyID, _ := jsonparser.GetInt(destObj, "Body")
		name := ""
		nameRaw, _ := jsonparser.GetString(destObj, "Name")
		if nameRaw != "" && !strings.HasPrefix(nameRaw, "$") {
			name = nameRaw
		} else {
			nameLoc, _ := jsonparser.GetString(destObj, "Name_Localised")
			if nameLoc != "" {
				name = nameLoc
			} else {
				name = nameRaw
			}
		}
		lastJournalState.Destination = Destination{
			SystemID: sysID,
			BodyID:   bodyID,
			Name:     name,
		}
	} else {
		lastJournalState.Destination = Destination{}
	}
}

// ParseJournalLine parses a single line of the journal and returns the new state after parsing.
func ParseJournalLine(line []byte, state *Journalstate) {
	re := regexp.MustCompile(`"event":"(\w*)"`)
	event := re.FindStringSubmatch(string(line))
	if len(event) < 2 {
		// Not a valid event line, skip
		return
	}
	p := parser{line}
	switch event[1] {
	case "Location":
		eLocation(p, state)
	case "SupercruiseEntry":
		eSupercruiseEntry(p, state)
	case "SupercruiseExit":
		eSupercruiseExit(p, state)
	case "FSDJump":
		eFSDJump(p, state)
	case "Touchdown":
		eTouchDown(p, state)
	case "Liftoff":
		eLiftoff(p, state)
	case "FSDTarget":
		eFSDTarget(p, state)
	case "ApproachBody":
		eApproachBody(p, state)
	case "ApproachSettlement":
		eApproachSettlement(p, state)
	case "Loadout":
		eLoadout(p) // NEW: handle Loadout event
	}
}

func eLocation(p parser, state *Journalstate) {
	// clear current location completely
	state.Type = LocationSystem
	state.Location.SystemAddress, _ = p.getInt(systemaddress)
	state.StarSystem, _ = p.getString(starsystem)

	bodyType, ok := p.getString(bodytype)

	if ok && bodyType == "Planet" {
		state.Location.BodyID, _ = p.getInt(bodyid)
		state.Location.Body, _ = p.getString(body)
		state.BodyType, _ = p.getString(bodytype)
		state.Type = LocationPlanet

		lat, ok := p.getFloat(latitude)
		if ok {
			state.Latitude = lat
			state.Longitude, _ = p.getFloat(longitude)
			state.Type = LocationLanded
		}
	}

	docked, _ := p.getBool(docked)
	if docked {
		state.Type = LocationDocked
	}
}

func eSupercruiseEntry(p parser, state *Journalstate) {
	state.Type = LocationSystem // don't throw away info
}

func eSupercruiseExit(p parser, state *Journalstate) {
	eLocation(p, state)
}

func eFSDJump(p parser, state *Journalstate) {
	eLocation(p, state)
}

func eTouchDown(p parser, state *Journalstate) {
	state.Latitude, _ = p.getFloat(latitude)
	state.Longitude, _ = p.getFloat(longitude)
	state.Type = LocationLanded
}

func eLiftoff(p parser, state *Journalstate) {
	state.Type = LocationPlanet
}

func eFSDTarget(p parser, state *Journalstate) {
	state.EDSMTarget.SystemAddress, _ = p.getInt(systemaddress)
	state.EDSMTarget.Name, _ = p.getString(name)

}

func eApproachBody(p parser, state *Journalstate) {
	state.Location.Body, _ = p.getString(body)
	state.Location.BodyID, _ = p.getInt(bodyid)

	state.Type = LocationPlanet
}

func eApproachSettlement(p parser, state *Journalstate) {
	state.Location.Body, _ = p.getString(bodyname)
	state.Location.BodyID, _ = p.getInt(bodyid)

	state.Type = LocationPlanet
}

func eLoadout(p parser) {
	capacity, ok := p.getInt("CargoCapacity")
	if ok {
		currentCargoCapacity = int(capacity)
	}
}

func getDisplayName(line []byte) string {
	name, _ := jsonparser.GetString(line, "Name")
	if name != "" && !strings.HasPrefix(name, "$") {
		return name
	}
	nameLoc, _ := jsonparser.GetString(line, "Name_Localised")
	if nameLoc != "" {
		return nameLoc
	}
	return name // fallback, even if it's a translation key
}
