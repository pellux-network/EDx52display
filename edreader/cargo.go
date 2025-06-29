package edreader

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/peterbn/EDx52display/mfd"
	log "github.com/sirupsen/logrus"
)

const FileCargo = "Cargo.json"

const (
	nameFileFolder        = "./names/"
	commodityNameFile     = nameFileFolder + "commodity.csv"
	rareCommodityNameFile = nameFileFolder + "rare_commodity.csv"
)

type Cargo struct {
	Count     int
	Inventory []CargoLine
}

type CargoLine struct {
	Name          string
	Count         int
	Stolen        int
	NameLocalized string `json:"Name_Localised"`
}

func (cl CargoLine) displayname() string {
	name := cl.Name
	displayName, ok := names[strings.ToLower(name)]
	if ok {
		name = displayName
	}
	return name
}

var (
	names        map[string]string
	currentCargo Cargo
)

func init() {
	log.Debugln("Initializing cargo name map...")
	initNameMap()
}

func RenderCargoPage(page *mfd.Page, _ Journalstate) {
	// If currentCargo is nil (never loaded), show "No cargo data"
	if currentCargo.Inventory == nil {
		page.Add("No cargo data")
		return
	}
	page.Add("#Cargo: %03d/%03d#", currentCargo.Count, ModulesInfoCargoCapacity())
	if len(currentCargo.Inventory) == 0 {
		page.Add("Cargo Hold Empty")
		return
	}
	sort.Slice(currentCargo.Inventory, func(i, j int) bool {
		a := currentCargo.Inventory[i]
		b := currentCargo.Inventory[j]
		return a.displayname() < b.displayname()
	})

	for _, line := range currentCargo.Inventory {
		page.Add("%s: %d", line.displayname(), line.Count)
	}
}

func handleCargoFile(file string) {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Debugln("No cargo file found:", file)
		currentCargo = Cargo{}
		return
	}
	var cargo Cargo
	json.Unmarshal(data, &cargo)
	currentCargo = cargo
}

func initNameMap() {
	commodity := readCsvFile(commodityNameFile)
	rareCommodity := readCsvFile(rareCommodityNameFile)

	names = make(map[string]string)

	mapCommodities := func(comms [][]string) {
		for _, com := range comms[1:] { //skipping the header line
			symbol := com[1]
			symbol = strings.ToLower(symbol)
			name := com[3]
			names[symbol] = name
		}
	}
	mapCommodities(commodity)
	mapCommodities(rareCommodity)
}

func readCsvFile(filename string) [][]string {
	csvfile, err := os.Open(filename)
	if err != nil {
		log.Panicln(err)
		return nil
	}
	defer csvfile.Close()
	csvreader := csv.NewReader(csvfile)
	records, err := csvreader.ReadAll()
	if err != nil {
		log.Panicln(err)
		return nil
	}
	return records
}
