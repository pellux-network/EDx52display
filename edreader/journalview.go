package edreader

import (
	"fmt"
	"log"
	"sort"
	"strings"

	lcdformat "github.com/pbxx/goLCDFormat"
	"github.com/pellux-network/EDx52display/edsm"
	"github.com/pellux-network/EDx52display/mfd"
)

// Gets system body information from EDSM
func GetEDSMBodies(systemaddress int64) (*edsm.System, error) {
	sysinfo := <-edsm.GetSystemBodies(systemaddress)
	if sysinfo.Error != nil {
		return nil, fmt.Errorf("unable to fetch system information: %w", sysinfo.Error)
	}
	sys := sysinfo.S
	if sys.ID64 == 0 {
		return nil, fmt.Errorf("no EDSM data for system address %d", systemaddress)
	}
	return &sys, nil
}

// Gets system monetary values from EDSM
func GetEDSMSystemValue(systemaddress int64) (*edsm.System, error) {
	valueinfopromise := edsm.GetSystemValue(systemaddress)
	valinfo := <-valueinfopromise
	if valinfo.Error != nil {
		return nil, fmt.Errorf("unable to fetch system value: %w", valinfo.Error)
	}
	return &valinfo.S, nil
}

// Page rendering functions for MFD
func RenderLocationPage(page *mfd.Page, state Journalstate) {
	if state.Type == LocationPlanet || state.Type == LocationLanded {
		ApplyBodyPage(page, "GRND SYS", &state)
	} else {
		ApplySystemPage(page, "CUR SYSTEM", state.Location.StarSystem, state.Location.SystemAddress, &state)
	}
}

func RenderDestinationPage(page *mfd.Page, state Journalstate) {
	// Show splash screen if enabled
	if state.ShowSplashScreen {
		page.Add("################")
		page.Add("EDx52display v0.2.0")
		page.Add("################")
		return
	}
	// Show arrival page if arrived at FSD target
	if state.ArrivedAtFSDTarget {
		page.Add("################")
		page.Add("  You have arrived  ")
		page.Add("################")
		return
	}
	// Show local destination if set, else FSD target, else "No Destination"
	if state.Destination.SystemAddress != 0 &&
		state.Destination.SystemAddress == state.Location.SystemAddress &&
		state.Destination.BodyID != 0 {

		ApplyBodyPage(page, "LOCAL TGT", &state)
	} else if state.EDSMTarget.SystemAddress != 0 {
		// header := fmt.Sprintf("FSD Target: %d", state.EDSMTarget.RemainingJumpsInRoute)
		ApplySystemPage(page, "NEXT JUMP", state.EDSMTarget.Name, state.EDSMTarget.SystemAddress, &state)
	} else {
		page.Add(" No Destination ")
	}
}

func RenderCargoPage(page *mfd.Page, _ Journalstate) {
	lines := []string{}
	// Cargo header
	lines = append(lines, fmt.Sprintf("CARGO: %04d/%04d", currentCargo.Count, ModulesInfoCargoCapacity()))
	// If currentCargo is nil (never loaded), show "No cargo data"
	if currentCargo.Inventory == nil {
		lines = append(lines, lcdformat.FillAround(16, "*", " NO CRGO DATA "))
		for _, line := range lines {
			page.Add(line)
		}
		return
	}

	if len(currentCargo.Inventory) == 0 {
		// If cargo inventory is empty, show "Cargo Hold Empty"
		lines = append(lines, lcdformat.FillAround(16, "*", " NO CARGO "))
		for _, line := range lines {
			page.Add(line)
		}
		return
	}
	sort.Slice(currentCargo.Inventory, func(i, j int) bool {
		a := currentCargo.Inventory[i]
		b := currentCargo.Inventory[j]
		return a.displayname() < b.displayname()
	})

	for _, line := range currentCargo.Inventory {
		lines = append(lines, lcdformat.SpaceBetween(16, line.displayname(), printer.Sprintf("%d", line.Count)))
	}
	// Add all pages in slice to the MFD
	for _, line := range lines {
		page.Add(line)
	}
}

// Page assembly functions for MFD
func ApplySystemPage(page *mfd.Page, header, systemname string, systemaddress int64, state *Journalstate) {
	// Initialize a slice to hold lines for the page
	lines := []string{}
	// Fetch system body information
	sys, err := GetEDSMBodies(systemaddress)
	if err != nil {
		log.Println("Error fetching EDSM data: ", err)
		return
	}

	// Fetch system monetary values
	values, err := GetEDSMSystemValue(systemaddress)
	if err != nil {
		log.Println("Error fetching EDSM system value: ", err)
		return
	}

	mainBody := sys.MainStar()
	// Separate the header (classification) and header display
	newHeader := header
	// Format the header based on the header title
	if header == "NEXT JUMP" || header == "CUR SYSTEM" {
		// Add FUEL indicator if star is scoopable
		if mainBody.IsScoopable {

			newHeader = lcdformat.SpaceBetween(16, header, "FUEL")
			lines = append(lines, newHeader)
		} else {
			lines = append(lines, header)
		}

	} else {
		lines = append(lines, header)
	}
	// Add the system name line to the page
	lines = append(lines, systemname)

	// Add the star class and remaining jumps
	// page.Add("Star: %s", mainBody.SubType)
	starTypeData := ParseStarTypeString(mainBody.SubType)
	jumps := ""

	if state != nil && header == "NEXT JUMP" {
		jumps = fmt.Sprintf("J:%d", state.EDSMTarget.RemainingJumpsInRoute)
	}
	lines = append(lines, lcdformat.SpaceBetween(16, fmt.Sprintf("CLS:%s", starTypeData.Class), jumps))
	// Add the main star information
	lines = append(lines, starTypeData.Desc)
	// Add system body count and estimated values

	lines = append(lines, lcdformat.SpaceBetween(16, "Bodies:", printer.Sprintf("%d", sys.BodyCount)))
	lines = append(lines, lcdformat.SpaceBetween(16, "Scan:", printer.Sprintf("%dcr", values.EstimatedValue)))
	lines = append(lines, lcdformat.SpaceBetween(16, "Map:", printer.Sprintf("%dcr", values.EstimatedValueMapped)))

	// Print valuable bodies if available
	if len(values.ValuableBodies) > 0 {
		lines = append(lines, lcdformat.FillAround(16, "*", " VAL BODIES "))
		for _, valbody := range values.ValuableBodies {
			bodyName := valbody.ShortName(*sys)
			crValue := printer.Sprintf("%dcr", valbody.ValueMax)
			// append the body name and value to the lines
			lines = append(lines, lcdformat.SpaceBetween(16, bodyName, crValue))
		}
	}

	// Evaluate presence of landable bodies and materials
	landables := []edsm.Body{}
	matLocations := map[string][]edsm.Body{}
	// Iterate through bodies to find landable bodies and their materials
	for _, body := range sys.Bodies {
		if body.IsLandable {
			landables = append(landables, body)
			for material := range body.Materials {
				bodiesWithMat, ok := matLocations[material]
				if !ok {
					bodiesWithMat = []edsm.Body{}
					matLocations[material] = bodiesWithMat
				}
				matLocations[material] = append(bodiesWithMat, body)
			}
		}
	}

	// Add prospecting information if landable bodies are present
	// if len(landables) > 0 {
	// 	lines = append(lines, lcdformat.FillAround(16, "*", " PROSPECT "))
	// 	materialList := []string{}

	// 	for mat := range matLocations {
	// 		materialList = append(materialList, mat)
	// 		bodies := matLocations[mat]
	// 		sort.Slice(bodies, func(i, j int) bool { return bodies[i].Materials[mat] > bodies[j].Materials[mat] })
	// 	}

	// 	// Sort materials by the number of bodies and then by material percentage
	// 	sort.Slice(materialList, func(i, j int) bool {
	// 		matA := materialList[i]
	// 		matB := materialList[j]
	// 		a := matLocations[matA]
	// 		b := matLocations[matB]
	// 		if len(a) == len(b) {
	// 			return a[0].Materials[matA] > b[0].Materials[matB]
	// 		}
	// 		return len(a) > len(b)

	// 	})
	// 	// Add material information to the page
	// 	for _, material := range materialList {
	// 		bodiesWithMat := matLocations[material]
	// 		lines = append(lines, fmt.Sprintf("%s %d", material, len(bodiesWithMat)))
	// 		b := bodiesWithMat[0]
	// 		// Add the body name (number usually) and material percentage
	// 		// matLine := lcdformat.SpaceBetween(16, b.ShortName(*sys), fmt.Sprintf("%.2f%%", float64(b.Materials[material])))
	// 		matLine := lcdformat.SpaceBetween(16, b.ShortName(*sys), fmt.Sprintf("%.2f%%%%", b.Materials[material]))
	// 		lines = append(lines, matLine)
	// 	}
	// } else {
	// 	return
	// }

	// Add all pages in slice to the MFD
	for _, line := range lines {
		page.Add(line)
	}
}

func ApplyBodyPage(page *mfd.Page, header string, state *Journalstate) {
	// page, "", state.Destination.Name, state.Location.SystemAddress, state.Destination.BodyID
	lines := []string{}

	sys, err := GetEDSMBodies(state.Location.SystemAddress)
	if err != nil {
		log.Println("Error fetching EDSM data: ", err)
		lines = append(lines, lcdformat.FillAround(16, "*", " EDSM ERROR "))
		return
	}

	body := sys.BodyByID(state.Destination.BodyID)
	if body.BodyID == 0 {
		lines = append(lines, lcdformat.FillAround(16, "*", " NO BODY DATA "))
		return
	}
	lines = append(lines, lcdformat.SpaceBetween(16, header, fmt.Sprintf("%.2fG", body.Gravity)))
	lines = append(lines, state.Destination.Name)
	lines = append(lines, body.SubType)

	// add the planet materials
	lines = append(lines, lcdformat.FillAround(16, "*", " MATERIAL "))
	for _, m := range body.MaterialsSorted() {
		lines = append(lines, lcdformat.SpaceBetween(16, fmt.Sprintf("%5.2f%%%%", m.Percentage), m.Name))
	}
	// Add all pages in slice to the MFD
	for _, line := range lines {
		page.Add(line)
	}
}

type StarTypeData struct {
	Class string
	Desc  string
}

func ParseStarTypeString(starType string) StarTypeData {
	// Parse the star type string and return a formatted version
	// Example input: K (Yellow-Orange) Star
	splitST := strings.Split(starType, " ")
	class := splitST[0]
	description := strings.ReplaceAll(splitST[1], "(", "")
	description = strings.ReplaceAll(description, ")", "")
	description = fmt.Sprintf("%s %s", description, "Star")
	return StarTypeData{
		Class: class,
		Desc:  description,
	}
}
