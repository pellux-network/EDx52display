# EDx52display

Reading Elite: Dangerous journal information and displaying on a Logitech X52 PRO MFD.

Please note that this software only works with the X52 Pro. The regular X52 HOTAS does not support third-party software for the MFD.

**NOTE 1: This fork is an attempt to update and improve the original app which has been abandoned. Currently, it replaces the parsing function to read 
line by line which is not only much more efficient but is working correctly again and doesn't "freeze". It also adds the ability to read the current
in-system destination and display EDSM info about it. Additionally, it includes another fork which improves logging.**

**NOTE 2: It is recommended to run a tool that uploads data to EDSM, such as [ED Market Connector](https://github.com/Marginal/EDMarketConnector). <br>
Doing this will ensure that any new discoveries can be shown on the display.**

## Installation

Simply download the latest release zip from the [releases](https://github.com/pellux-network/EDx52display/releases/latest) page or build the app yourself with
`go build -o EDx52Display.exe`

## Output

Running this application will show 3 pages of information on your MFD. Most of this information is sourced from EDSM.net.

Of particular note is:

- Live view of cargo hold - *keep track while mining*
- Value of scanning and mapping the system - *know where to go, without checking system map*
- Surface gravity of the planet you are about to land on - *avoid becoming a stellar pancake!*

### Page 1: Destination

This page adapts based on your current target. If an in-system destination is selected, the "System Destination" page will be displayed, showing the name of the celestial body along with 
relevant information from EDSM, such as its gravity. If the target is another system, the "FSD Destination" page will appear, providing the name of the next jump's system, details about 
whether the star is scoopable, and additional system information.

### Page 2: Current Location

This page provides details about your current location, which may refer to the system you are situated in or the planet you have approached. 
It also includes additional information regarding the location.

### Page 3: Cargo

This page displays the current occupancy of your cargo hold along with a detailed list of specific items and their respective quantities.



### System Page

A page with system information will have the following information, sourced from EDSM:

- System Name
- Whether the main star is scoopable
- Number of bodies (as reported by EDSM)
- Total value for scanning the system
- Total value for mapping the entire system
- Any valuable bodies
- System Prospecting information
  - Available elements, with number of planets landable where they occur
  - The planet in the system with the highest occurence of said element

### Planet Page

A page with planet information will have the following data, sourced from EDSM:

- Planet name
- Planet Gravity (!)
- Available materials for the planet, if any

## Buttons / Navigation

This tool will use both function wheels on the MFD.

The left wheel will scroll between pages

The right wheel will scroll a page up and down

**Pressing** the right wheel will refresh data from EDSM. The display will cache values from EDSM to avoid hitting their API rate limit. 
Pressing this button will update with new data, which is useful if you have recently scanned the system and uploaded data with ED Market Connector or similar tools.

## Troubleshooting

This application reads the journal files of your elite dangerous installation.
These are normally located in `%USERPROFILE%\\Saved Games\\Frontier Developments\\Elite Dangerous` on Windows. However, if your installation
uses a different location, you should update the conf.yaml file in the installation folder.

### Command Line Arguments

- `--log`: Set the desired log level. One of:
  - panic 
  - fatal 
  - error
  - warning
  - info (default)
  - debug 
  - trace

## Credits

This project owes a great deal to [Anthony Zaprzalka](https://github.com/AZaps) in terms of idea and execution
and to [Jonathan Harris](https://github.com/Marginal) and the [EDMarketConnector](https://github.com/Marginal/EDMarketConnector) project
for the CSV files of names for all the commodities.
