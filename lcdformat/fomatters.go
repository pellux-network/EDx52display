package lcdformat

import (
	"fmt"
	"strings"
)

func NRawChar(numSpaces int, fillChar string) string {
	var spaces string
	for i := 0; i < numSpaces; i++ {
		spaces += fillChar
	}
	return spaces
}

func SpaceBetween(strings []string, width int) string {
	return FillBetween(strings, width, " ")
}

func FillBetween(strings []string, width int, fillChar string) string {
	var out string

	// Calculate the total length of all strings
	totalLength := 0
	for _, str := range strings {
		totalLength += len(str)
	}
	// Calculate total number of spaces needed
	totalSpaces := width - totalLength
	if totalSpaces <= 0 {
		// If total spaces is negative, append each string without spaces
		for _, str := range strings {
			out += str
		}
	} else {
		// Calculate spaces to insert between each string
		numGaps := len(strings) - 1
		if numGaps == 0 {
			// Only one string, pad with fillChar to the right
			out = strings[0] + NRawChar(totalSpaces, fillChar)
		} else {
			spacesPerGap := totalSpaces / numGaps
			extraSpaces := totalSpaces % numGaps
			for i, str := range strings {
				out += str
				if i < numGaps {

					numSpaces := spacesPerGap
					// Distribute the remainder (extraSpaces) to the rightmost gaps
					if i >= numGaps-extraSpaces {
						numSpaces++
					}
					out += NRawChar(numSpaces, fillChar)
				}
			}
		}
	}
	// Trim the output to ensure it does not exceed the specified width
	// if len(out) > width {
	// 	out = out[:width]
	// }
	return out
}

func Center(str string, width int) string {
	// Center the string within the specified width using spaces
	return Surround(str, width, " ")
}

func Surround(str string, width int, fillChar string) string {
	// Center the string within the specified width using fillChar
	if len(str) >= width {
		return str
	}
	totalSpaces := width - len(str)
	leftSpaces := totalSpaces / 2
	rightSpaces := totalSpaces - leftSpaces
	return NRawChar(leftSpaces, fillChar) + str + NRawChar(rightSpaces, fillChar)
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
