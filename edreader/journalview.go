package edreader

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/peterbn/EDx52display/edsm"
	"github.com/peterbn/EDx52display/mfd"
)

func addValueRight(page *mfd.Page, label string, value int64) {
	valstr := fmt.Sprintf("%dcr", value)
	pad := 16 - (len(label) + 1 + len(valstr))
	if pad < 0 {
		pad = 0
	}
	page.Add("%s:%s%s", label, strings.Repeat(" ", pad), valstr)
}

func addIntRight(page *mfd.Page, label string, value int) {
	valstr := fmt.Sprintf("%d", value)
	pad := 16 - (len(label) + 1 + len(valstr))
	if pad < 0 {
		pad = 0
	}
	page.Add("%s:%s%s", label, strings.Repeat(" ", pad), valstr)
}

func RenderLocationPage(page *mfd.Page, state Journalstate) {
	if state.Type == LocationPlanet || state.Type == LocationLanded {
		RenderEDSMBody(page, "#    Planet    #", state.Location.Body, state.Location.SystemAddress, state.Location.BodyID)
	} else {
		RenderEDSMSystem(page, "#    System    #", state.Location.StarSystem, state.Location.SystemAddress)
	}
}

func RenderFSDTarget(page *mfd.Page, state Journalstate) {
	if state.EDSMTarget.SystemAddress == 0 {
		page.Add("No FSD Target")
	} else {
		RenderEDSMSystem(page, "#  FSD Target  #", state.EDSMTarget.Name, state.EDSMTarget.SystemAddress)
	}
}

func RenderDestinationPage(page *mfd.Page, state Journalstate) {
	// Show local destination if set, else FSD target, else "No Destination"
	if state.Destination.SystemID != 0 &&
		state.Destination.SystemID == state.Location.SystemAddress &&
		state.Destination.BodyID != 0 {
		page.Add("# Local Target #")
		page.Add(state.Destination.Name)
		RenderEDSMBody(page, "", state.Destination.Name, state.Location.SystemAddress, state.Destination.BodyID)
	} else if state.EDSMTarget.SystemAddress != 0 {
		RenderEDSMSystem(page, "#  FSD Target  #", state.EDSMTarget.Name, state.EDSMTarget.SystemAddress)
	} else {
		page.Add("No Destination")
	}
}

func RenderEDSMSystem(page *mfd.Page, header, systemname string, systemaddress int64) {
	sysinfopromise := edsm.GetSystemBodies(systemaddress)
	valueinfopromise := edsm.GetSystemValue(systemaddress)

	page.Add(header)
	page.Add(systemname)

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

	mainBody := sys.MainStar()
	if mainBody.IsScoopable {
		page.Add("Scoopable")
	} else {
		page.Add("Not scoopable")
	}

	page.Add(mainBody.SubType)

	addIntRight(page, "Bodies", sys.BodyCount)

	valinfo := <-valueinfopromise
	if valinfo.Error == nil {
		addValueRight(page, "Scan", valinfo.S.EstimatedValue)
		addValueRight(page, "Map", valinfo.S.EstimatedValueMapped)

		if len(valinfo.S.ValuableBodies) > 0 {
			page.Add("Valuable Bodies:")
		}
		for _, valbody := range valinfo.S.ValuableBodies {
			bname := valbody.ShortName(sys)
			valstr := printer.Sprintf("%dcr", valbody.ValueMax)
			pad := 1
			if len(bname)+len(valstr) < 16 {
				pad = 16 - (len(bname) + len(valstr))
			}
			padstr := strings.Repeat(" ", pad)
			page.Add("%s%s%s", bname, padstr, valstr)
		}
	}

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

	if len(landables) == 0 {
		return
	}

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
		page.Add("%s: %.2f%%", b.ShortName(sys), b.Materials[mat])
	}
}

func RenderEDSMBody(page *mfd.Page, header, bodyName string, systemaddress, bodyid int64) {
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
