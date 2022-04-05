package ansiart2utf8

type OC struct {
	Hex      string
	Xterm256 int
}

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

func TranslateColors(sSGR []int, bIntense bool) []int {

	sRet := make([]int, 0, len(sSGR))

	for _, v := range sSGR {

		// FOREGROUND COLORS
		if IsBtween(v, 30, 37) {

			if bIntense {
				sRet = append(sRet, 38, 5, OrigLight[v-30].Xterm256)
			} else {
				sRet = append(sRet, 38, 5, OrigDark[v-30].Xterm256)
			}
			continue
		}

		if IsBtween(v, 90, 97) {
			sRet = append(sRet, 38, 5, OrigLight[v-90].Xterm256)
			continue
		}

		// BACKGROUND COLORS
		if IsBtween(v, 40, 47) {
			sRet = append(sRet, 48, 5, OrigDark[v-40].Xterm256)
			continue
		}

		if IsBtween(v, 100, 107) {
			sRet = append(sRet, 48, 5, OrigLight[v-100].Xterm256)
			continue
		}

		sRet = append(sRet, v)
	}

	return sRet
}
