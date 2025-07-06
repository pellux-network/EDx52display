package edreader

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/pellux-network/EDx52display/edsm"
	"github.com/pellux-network/EDx52display/lcdformat"
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
		ApplyBodyPage(page, "GRND SYS", state.Location.Body, state.Location.SystemAddress, state.Location.BodyID)
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
		page.Add("LOCAL TGT")
		page.Add(state.Destination.Name)
		ApplyBodyPage(page, "", state.Destination.Name, state.Location.SystemAddress, state.Destination.BodyID)
	} else if state.EDSMTarget.SystemAddress != 0 {
		// header := fmt.Sprintf("FSD Target: %d", state.EDSMTarget.RemainingJumpsInRoute)
		header := fmt.Sprintf("NEXT JUMP")

		ApplySystemPage(page, header, state.EDSMTarget.Name, state.EDSMTarget.SystemAddress, &state)
	} else {
		page.Add(" No Destination ")
	}
}

// Page assembly functions for MFD
func ApplySystemPage(page *mfd.Page, header, systemname string, systemaddress int64, state *Journalstate) {
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
	newHeader := header
	// Format the header based on the header title
	if header == "NEXT JUMP" || header == "CUR SYSTEM" {
		// Add FUEL indicator if star is scoopable
		if mainBody.IsScoopable {
			newHeader = lcdformat.SpaceBetween([]string{header, "FUEL"}, 16)
			page.Add(newHeader)
		} else {
			page.Add(header)
		}

	} else {
		page.Add(header)
	}
	// Add the system name line to the page
	page.Add(systemname)

	// Add the star class and remaining jumps
	// page.Add("Star: %s", mainBody.SubType)
	starTypeData := lcdformat.ParseStarTypeString(mainBody.SubType)
	jumps := ""

	if state != nil && header == "NEXT JUMP" {
		jumps = fmt.Sprintf("J:%d", state.EDSMTarget.RemainingJumpsInRoute)
	}
	page.Add(lcdformat.SpaceBetween([]string{
		fmt.Sprintf("CLS:%s", starTypeData.Class),
		jumps,
	}, 16))
	// Add the main star information
	page.Add(starTypeData.Desc)
	// Add system body count and estimated values
	page.Add(lcdformat.SpaceBetween([]string{"Bodies:", printer.Sprintf("%d", sys.BodyCount)}, 16))
	page.Add(lcdformat.SpaceBetween([]string{"Scan:", printer.Sprintf("%dcr", values.EstimatedValue)}, 16))
	page.Add(lcdformat.SpaceBetween([]string{"Map:", printer.Sprintf("%dcr", values.EstimatedValueMapped)}, 16))

	// Print valuable bodies if available
	if len(values.ValuableBodies) > 0 {
		page.Add("Valuable Bodies:")
		for _, valbody := range values.ValuableBodies {
			bname := valbody.ShortName(*sys)
			valstr := printer.Sprintf("%dcr", valbody.ValueMax)
			pad := 1
			if len(bname)+len(valstr) < 16 {
				pad = 16 - (len(bname) + len(valstr))
			}
			padstr := strings.Repeat(" ", pad)
			page.Add("%s%s%s", bname, padstr, valstr)
		}
	}

	// Evaluate presence of landable bodies and materials
	landables := []edsm.Body{}
	matLocations := map[string][]edsm.Body{}

	for _, b := range sys.Bodies {
		if b.IsLandable {
			landables = append(landables, b)
			for m := range b.Materials {
				mlist, ok := matLocations[m]
				if !ok {
					mlist = []edsm.Body{}
					matLocations[m] = mlist
				}
				matLocations[m] = append(mlist, b)
			}
		}
	}

	// Add prospecting information if landable bodies are present
	if len(landables) > 0 {
		page.Add("Prospecting:")
		matlist := []string{}
		for mat := range matLocations {
			matlist = append(matlist, mat)
			bodies := matLocations[mat]
			sort.Slice(bodies, func(i, j int) bool { return bodies[i].Materials[mat] > bodies[j].Materials[mat] })
		}

		sort.Slice(matlist, func(i, j int) bool {
			matA := matlist[i]
			matB := matlist[j]
			a := matLocations[matA]
			b := matLocations[matB]
			if len(a) == len(b) {
				return a[0].Materials[matA] > b[0].Materials[matB]
			}
			return len(a) > len(b)

		})
		for _, mat := range matlist {
			bodies := matLocations[mat]
			page.Add("%s %d", mat, len(bodies))
			b := bodies[0]
			page.Add("%s: %.2f%%", b.ShortName(*sys), b.Materials[mat])
		}
	} else {
		return
	}

}

func ApplyBodyPage(page *mfd.Page, header, bodyName string, systemaddress, bodyid int64) {
	sysinfopromise := edsm.GetSystemBodies(systemaddress)
	page.Add(header)
	page.Add(bodyName)
	sysinfo := <-sysinfopromise
	if sysinfo.Error != nil {
		log.Println("Unable to fetch system information: ", sysinfo.Error)
		page.Add("Sysinfo lookup error")
		return
	}
	sys := sysinfo.S
	if sys.ID64 == 0 {
		page.Add("No EDSM data")
		return
	}

	body := sys.BodyByID(bodyid)
	if body.BodyID == 0 {
		page.Add("No EDSM data")
		return
	}

	page.Add("Gravity %7.2fG", body.Gravity)

	page.Add("Materials:")
	for _, m := range body.MaterialsSorted() {
		page.Add("%5.2f%% %s", m.Percentage, m.Name)
	}
}
