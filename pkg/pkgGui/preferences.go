package pkgGui

import (
	"log"
	"strings"
)

const (
	PREF_LIMIT_FIRST  = "€m10x.de€"
	PREF_LIMIT_SECOND = "$m10x.de$"
)

func dataToString() string {
	strArr := ""
	for i, d := range data {
		if i == 0 { // skip first row which is the header row
			continue
		}
		for ii, dd := range d {
			if ii > 0 {
				strArr += PREF_LIMIT_FIRST
			}
			strArr += dd
		}
		strArr += PREF_LIMIT_SECOND
	}
	return strArr
}

func stringToData(str string) {
	for _, d := range strings.Split(str, PREF_LIMIT_SECOND) {
		dataRow := strings.Split(d, PREF_LIMIT_FIRST)
		if len(dataRow) == COLUMN_COUNT { // empty or too short dataRow will cause fyne to crash
			data = append(data, dataRow)
		}
	}
}

func setPreferences(versionTool string) {
	log.Println("saving preferences")
	a.Preferences().SetString("version", versionTool)
	a.Preferences().SetString("device", comboSelect.Selected)
	a.Preferences().SetString("data", dataToString())
}

func getPreferences(versionTool string) {
	versionPref := a.Preferences().StringWithFallback("version", "none")
	if versionPref == "none" {
		log.Println("No Preferences found")
		return
	} else if versionPref != versionTool {
		log.Printf("Version of preferences (%s) may not be compatible with version of midi2key-ng (%s)\n", versionPref, versionTool)
	} else {
		log.Printf("Loading preferences...")
	}

	deviceOptions := comboSelect.Options
	prefDevice := a.Preferences().StringWithFallback("device", "")
	for i := 0; i < len(deviceOptions); i++ {
		if deviceOptions[i] == prefDevice {
			comboSelect.SetSelectedIndex(i)
		}
	}
	prefData := a.Preferences().StringWithFallback("data", "")
	stringToData(prefData)
	log.Printf("Preferences loaded")
}
