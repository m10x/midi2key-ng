package pkgControllers

import (
	"fmt"
)

type Controller struct {
	Name string
	Data [][]string
}

var Controllers = []Controller{
	{
		Name: "X-Touch Mini",
		Data: [][]string{{"S8", "-", "-", "16256", "false"}, {"K16", "-", "-", "33", "false"}, {"B32", "-", "-", "255", "false"}, {"K17", "-", "-", "33", "false"}, {"B33", "-", "-", "255", "false"}, {"K18", "-", "-", "33", "false"}, {"B34", "-", "-", "255", "false"}, {"K19", "-", "-", "33", "false"}, {"B35", "-", "-", "255", "false"}, {"K20", "-", "-", "33", "false"}, {"B36", "-", "-", "255", "false"}, {"K21", "-", "-", "33", "false"}, {"B37", "-", "-", "255", "false"}, {"K22", "-", "-", "33", "false"}, {"B38", "-", "-", "255", "false"}, {"K23", "-", "-", "33", "false"}, {"B39", "-", "-", "255", "false"}, {"B89", "-", "-", "255", "false"}, {"B87", "-", "-", "255", "false"}, {"B90", "-", "-", "255", "false"}, {"B88", "-", "-", "255", "false"}, {"B40", "-", "-", "255", "false"}, {"B91", "-", "-", "255", "false"}, {"B41", "-", "-", "255", "false"}, {"B92", "-", "-", "255", "false"}, {"B42", "-", "-", "255", "false"}, {"B86", "-", "-", "255", "false"}, {"B43", "-", "-", "255", "false"}, {"B93", "-", "-", "255", "false"}, {"B44", "-", "-", "255", "false"}, {"B94", "-", "-", "255", "false"}, {"B45", "-", "-", "255", "false"}, {"B95", "-", "-", "255", "false"}, {"B84", "-", "-", "255", "false"}, {"B85", "-", "-", "255", "false"}},
	},
}

func PrintCurrentConfigAsCode(arr [][]string) string {
	var output string
	output += "Your current config is [][]string{"
	for i, inner := range arr {
		if i > 0 {
			output += ", "
		}
		output += "{"
		for j, str := range inner {
			if j > 0 {
				output += ", "
			}
			output += fmt.Sprintf("%q", str)
		}
		output += "}"
	}
	output += "}"

	return output
}
