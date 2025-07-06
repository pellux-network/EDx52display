package edreader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/pellux-network/EDx52display/mfd"
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
		addCargoRight(page, line.displayname(), line.Count)
	}
}

func addCargoRight(page *mfd.Page, label string, value int) {
	valstr := fmt.Sprintf("%d", value)
	pad := 16 - (len(label) + 1 + len(valstr))
	if pad < 0 {
		pad = 0
	}
	page.Add("%s:%s%s", label, strings.Repeat(" ", pad), valstr)
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

func mapCommodities(data [][]string, symbolIdx, nameIdx int) {
	for _, com := range data[1:] {
		symbol := com[symbolIdx]
		symbol = strings.ToLower(symbol)
		name := com[nameIdx]
		names[symbol] = name
	}
}

func initNameMap() {
	commodity := readCsvFile(commodityNameFile)
	rareCommodity := readCsvFile(rareCommodityNameFile)

	names = make(map[string]string)

	// commodity.csv: symbol at 1, name at 3
	mapCommodities(commodity, 1, 3)
	// rare_commodity.csv: symbol at 1, name at 4
	mapCommodities(rareCommodity, 1, 4)
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
