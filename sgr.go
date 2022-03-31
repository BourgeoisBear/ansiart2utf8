package ansiart2utf8

import (
	"fmt"
	"strings"
)

const (
	SGR_BOLD uint32 = 1 << iota
	SGR_FAINT
	SGR_ITALIC
	SGR_UNDERLINE
	SGR_BLNK_SLOW
	SGR_BLNK_FAST
	SGR_INVERSE
	SGR_CONCEAL
	SGR_STRIKETHROUGH
)

type SGR struct {
	// Bold, Faint, Italic, Underline, Blink, Inverse, Conceal, Strikethrough bool
	Flags    uint32
	ColorTxt string
	ColorBg  string
}

func (pS *SGR) Reset() {
	*pS = SGR{}
}

func (pS *SGR) Fset(f uint32) {
	pS.Flags |= f
}

func (pS *SGR) Fclr(f uint32) {
	pS.Flags &^= f
}

func (pS *SGR) IsEqual(pCodePrev *SGR) bool {

	return ((pS.Flags == pCodePrev.Flags) &&
		(pS.ColorTxt == pCodePrev.ColorTxt) &&
		(pS.ColorBg == pCodePrev.ColorBg))
}

func (pS *SGR) ToEsc(bFakeEscape bool) string {

	bsParts := []string{"0"}

	bsIter := []struct {
		Flag uint32
		Set  string
	}{
		{SGR_BOLD, "1"},
		{SGR_FAINT, "2"},
		{SGR_ITALIC, "3"},
		{SGR_UNDERLINE, "4"},
		{SGR_BLNK_SLOW, "5"},
		{SGR_BLNK_FAST, "6"},
		{SGR_INVERSE, "7"},
		{SGR_CONCEAL, "8"},
		{SGR_STRIKETHROUGH, "9"},
	}

	for _, sITER := range bsIter {

		if (sITER.Flag & pS.Flags) != 0 {
			bsParts = append(bsParts, sITER.Set)
		}
	}

	if len(pS.ColorBg) > 0 {
		bsParts = append(bsParts, pS.ColorBg)
	}

	if len(pS.ColorTxt) > 0 {
		bsParts = append(bsParts, pS.ColorTxt)
	}

	if len(bsParts) > 0 {

		pfx := "\x1b["
		if bFakeEscape {
			pfx = pfx + "96m^[" + pfx + "0m["
		}

		return pfx + strings.Join(bsParts, ";") + "m"
	}

	return ""
}

/*
	MergeCodes SGR int codes (like ESC[0m) into an existing SGR struct
*/
func (pS *SGR) MergeCodes(biCodes []int) error {

	type Action struct {
		Set   bool
		Flags uint32
	}

	ACT := map[int]Action{
		1:  Action{true, SGR_BOLD},
		2:  Action{true, SGR_FAINT},
		3:  Action{true, SGR_ITALIC},
		4:  Action{true, SGR_UNDERLINE},
		5:  Action{true, SGR_BLNK_SLOW},
		6:  Action{true, SGR_BLNK_FAST},
		7:  Action{true, SGR_INVERSE},
		8:  Action{true, SGR_CONCEAL},
		9:  Action{true, SGR_STRIKETHROUGH},
		21: Action{false, SGR_BOLD},
		22: Action{false, SGR_BOLD | SGR_FAINT},
		23: Action{false, SGR_ITALIC},
		24: Action{false, SGR_UNDERLINE},
		25: Action{false, SGR_BLNK_SLOW | SGR_BLNK_FAST},
		27: Action{false, SGR_INVERSE},
		28: Action{false, SGR_CONCEAL},
		29: Action{false, SGR_STRIKETHROUGH},
	}

	nCodes := len(biCodes)

	for i := 0; i < nCodes; i++ {

		switch biCodes[i] {

		// RESET
		case 0:
			pS.Reset()

		// DEFAULT FG
		case 39:
			pS.ColorTxt = ""

		// DEFAULT BG
		case 49:
			pS.ColorBg = ""

		// HIGH COLOR FG
		// HIGH COLOR BG
		case 38, 48:

			szColor, nAdvance := HighColor(biCodes[i:])

			if nAdvance > 0 {

				switch biCodes[i] {
				case 38:
					pS.ColorTxt = szColor
				case 48:
					pS.ColorBg = szColor
				}

				i += nAdvance
				continue

			} else {

				return fmt.Errorf("SGR-SKIP [HCOLOR]: %d", biCodes[i])
			}

		default:

			if oA, bOK := ACT[biCodes[i]]; bOK {

				// DISPLAY ATTRIBUTES (bold, underline, etc)
				if oA.Set {
					pS.Fset(oA.Flags)
				} else {
					pS.Fclr(oA.Flags)
				}

			} else if (biCodes[i] >= 30) && (biCodes[i] <= 37) {

				// CLASSIC FG
				pS.ColorTxt = fmt.Sprintf("%d", biCodes[i])

			} else if (biCodes[i] >= 40) && (biCodes[i] <= 47) {

				// CLASSIC BG
				pS.ColorBg = fmt.Sprintf("%d", biCodes[i])

			} else {

				return fmt.Errorf("SGR-SKIP [UNKWN]: %d", biCodes[i])
			}
		}
	}

	return nil
}

/*
	Formats high color SGR codes
*/
func HighColor(arCodes []int) (string, int) {

	nCodes := len(arCodes)

	fnKosher := func(i int) bool {

		return ((i >= 0) && (i <= 255))
	}

	if nCodes >= 3 {

		switch arCodes[1] {

		// 5;n where n is color index (0..255)
		case 5:

			if fnKosher(arCodes[2]) {

				return fmt.Sprintf("%d;%d;%d", arCodes[0], arCodes[1], arCodes[2]), 2
			}

		// 2;r;g;b where r,g,b are red, green and blue color channels (out of 255)
		case 2:

			if nCodes >= 5 {

				if fnKosher(arCodes[2]) && fnKosher(arCodes[3]) && fnKosher(arCodes[4]) {

					return fmt.Sprintf("%d;%d;%d;%d;%d",
						arCodes[0], arCodes[1], arCodes[2], arCodes[3], arCodes[4]), 4
				}
			}
		}
	}

	return "", 0
}
