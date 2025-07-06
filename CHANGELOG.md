# Changelog

## [v0.2.0] - XX-XX-XXXX [PLANNED RELEASE]

### Added
- Jumps remaining to FSD Target page
- Arrival screen when route is complete
- Loading splashscreen
- Support for the Panther Clipper's massive cargo hold by displaying 4 digits on the cargo screen

### Changed
- Polling to OS-level notifications, faster and more efficient
- Most value formatting to be right-aligned, may change more in future releases
- Credit value formatting to include commas in the appropriate places for better readability

### Fixed
- Target page sometimes displaying unlocalized name
- Rare commodities displaying the category of the commodity instead of the name
- Outdated commodity CSVs (May still be incomplete)

### Known Bugs
- Selecting a system from the left-side external panel results in either 0 or 16 jumps remaining being displayed. This is unfortunately a bug with ED's journaling where it's actually displaying those number as jumps remaining so this will require a fix on Frontier's end
- Arrival screen displays a few seconds after startup

## [v0.1.3] - 06-29-2025

### Added
- `-s -w` flags to strip debug info
- UPX compression
※ These changes result in the release executable dropping from about
   11MB to 2.4MB!!!

### Changed
- Default polling rate to 500ms

## [v0.1.2] - 06-29-2025

### Added
- Page registry and config-driven page toggling.

## [v0.1.1] - 06-29-2025

### Added
- Page registry and config-driven page toggling.
- System tray support with quit option.
- Logging to rotating files in the `logs` directory.
- Icon embedding for system tray and executable.
- Cargo page now shows "Cargo Hold Empty" when appropriate.

### Changed
- Destination page now dynamically shows local or FSD target.
- Logging format and file naming improved.

### Fixed
- Cargo page no longer shows "No cargo data" when cargo is empty.
- Fixed issues with duplicate function names and package imports.

## [v0.1.0] - 06-28-2025

### Added
- Initial fixes and journal parsing
