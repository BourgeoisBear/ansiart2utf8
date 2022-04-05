package ansiart2utf8

var OrigDark = []OC{
	OC{Hex: `#000000`, Xterm256: 16},
	OC{Hex: `#AB0000`, Xterm256: 124},
	OC{Hex: `#00AB00`, Xterm256: 34},
	OC{Hex: `#AB5700`, Xterm256: 130},
	OC{Hex: `#0000AB`, Xterm256: 19},
	OC{Hex: `#AB00AB`, Xterm256: 127},
	OC{Hex: `#00ABAB`, Xterm256: 37},
	OC{Hex: `#ABABAB`, Xterm256: 248},
}

var OrigLight = []OC{
	OC{Hex: `#575757`, Xterm256: 240},
	OC{Hex: `#FF5757`, Xterm256: 203},
	OC{Hex: `#57FF57`, Xterm256: 83},
	OC{Hex: `#FFFF57`, Xterm256: 227},
	OC{Hex: `#5757FF`, Xterm256: 63},
	OC{Hex: `#FF57FF`, Xterm256: 207},
	OC{Hex: `#57FFFF`, Xterm256: 87},
	OC{Hex: `#FFFFFF`, Xterm256: 15},
}

// TODO: what does overwrite do?
// (set color at start, put char at end, move to start, change color)
// NOTE: SGR sets pen, but cell not touched until painted
// TODO: make conditional
// TODO: color striping issues
// TODO: always treat bold as bright color
func TranslateColors(sSGR []int) []int {

	//return sSGR

	bIntense := false
	for _, val := range sSGR {

		switch val {

		case 1:
			bIntense = true

		case 0, 2, 22:
			bIntense = false
		}
	}

	sRet := make([]int, 0, len(sSGR))

	for _, v := range sSGR {

		// FOREGROUND COLORS
		if (v >= 30) && (v <= 37) {

			if bIntense {
				sRet = append(sRet, 38, 5, OrigLight[v-30].Xterm256)
			} else {
				sRet = append(sRet, 38, 5, OrigDark[v-30].Xterm256)
			}
			continue
		}

		if (v >= 90) && (v <= 97) {
			sRet = append(sRet, 38, 5, OrigLight[v-90].Xterm256)
			continue
		}

		// BACKGROUND COLORS
		if (v >= 40) && (v <= 47) {
			sRet = append(sRet, 48, 5, OrigDark[v-40].Xterm256)
			continue
		}

		if (v >= 100) && (v <= 107) {
			sRet = append(sRet, 48, 5, OrigLight[v-100].Xterm256)
			continue
		}

		sRet = append(sRet, v)
	}

	return sRet
}
