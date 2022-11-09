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
		if len(dataRow) == len(data[0]) { // empty or too short dataRow will cause fyne to crash
			data = append(data, dataRow)
		}
	}
}

func setPreferences(versionPref int) {
	log.Println("saving preferences")
	a.Preferences().SetInt("version", versionPref)
	a.Preferences().SetString("device", comboSelect.Selected)
	a.Preferences().SetString("data", dataToString())
}

func getPreferences(versionPref int) {
	prefVersion := a.Preferences().IntWithFallback("version", 0)
	if prefVersion == 0 {
		log.Println("No Preferences found")
		return
	} else if prefVersion != versionPref {
		log.Println("Incompatible Preferences Version")
		return
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
}
