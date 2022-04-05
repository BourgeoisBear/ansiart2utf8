package ansiart2utf8

import (
	"fmt"
	"strconv"
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

const (
	CIX_FG = iota
	CIX_BG
	CIX_MAX
)

const (
	DEFAULT_FG int = 37
	DEFAULT_BG int = 40
)

type SGR struct {
	// Bold, Faint, Italic, Underline, Blink, Inverse, Conceal, Strikethrough bool
	Flags uint32
	Color [CIX_MAX][]int
}

func (pS *SGR) Fset(f uint32) {
	pS.Flags |= f
}

func (pS *SGR) Fclr(f uint32) {
	pS.Flags &^= f
}

func IaEqual(A, B []int) bool {

	if len(A) != len(B) {
		return false
	}

	for ix := range A {
		if A[ix] != B[ix] {
			return false
		}
	}

	return true
}

func (pS *SGR) ToEsc(pPrev *SGR, bAsDiff, bFakeEscape bool) string {

	// TODO: brighten bold colors?
	sParts := []int{}

	bsIter := []struct {
		Flag  uint32
		Set   int
		Clear int
	}{
		{SGR_BOLD, 1, 22},
		{SGR_FAINT, 2, 22},
		{SGR_ITALIC, 3, 23},
		{SGR_UNDERLINE, 4, 24},
		{SGR_BLNK_SLOW, 5, 25},
		{SGR_BLNK_FAST, 6, 25},
		{SGR_INVERSE, 7, 27},
		{SGR_CONCEAL, 8, 28},
		{SGR_STRIKETHROUGH, 9, 29},
	}

	// APPEND ANSI CODES FOR TEXT STYLE
	flagDiff := pS.Flags ^ pPrev.Flags
	for _, sITER := range bsIter {

		// CLEAR
		if bAsDiff && ((sITER.Flag & flagDiff) != 0) {
			if (sITER.Flag & pS.Flags) != 0 {
				sParts = append(sParts, sITER.Set)
			} else {
				sParts = append(sParts, sITER.Clear)
			}
		}

		if (sITER.Flag & pS.Flags) != 0 {
			sParts = append(sParts, sITER.Set)
		}
	}

	// APPEND ANSI CODES FOR FG/BG COLORS
	mColor := map[int]int{
		CIX_FG: DEFAULT_FG,
		CIX_BG: DEFAULT_BG,
	}

	for CIX := range []int{CIX_FG, CIX_BG} {

		if bAsDiff && IaEqual(pS.Color[CIX], pPrev.Color[CIX]) {
			continue
		}

		if len(pS.Color[CIX]) == 0 {
			sParts = append(sParts, mColor[CIX])
		} else {
			sParts = append(sParts, pS.Color[CIX]...)
		}
	}

	// EARLY EXIT
	if len(sParts) == 0 {
		return ""
	}

	// GENERATE ESCAPE CODE
	pfx := "\x1b["
	if bFakeEscape {
		pfx = pfx + "96m^[" + pfx + "0m["
	}

	sStr := make([]string, len(sParts))
	for ix := range sParts {
		sStr[ix] = strconv.FormatInt(int64(sParts[ix]), 10)
	}
	return pfx + strings.Join(sStr, ";") + "m"
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
			*pS = SGR{}

		// DEFAULT FG
		case 39:
			pS.Color[CIX_FG] = []int{DEFAULT_FG}

		// DEFAULT BG
		case 49:
			pS.Color[CIX_BG] = []int{DEFAULT_BG}

		// HIGH COLOR FG
		// HIGH COLOR BG
		case 38, 48:

			sColor, nAdvance := HighColor(biCodes[i:])

			if nAdvance > 0 {

				switch biCodes[i] {
				case 38:
					pS.Color[CIX_FG] = sColor
				case 48:
					pS.Color[CIX_BG] = sColor
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

			} else if isBtween(biCodes[i], 30, 37) || isBtween(biCodes[i], 90, 97) {

				// CLASSIC FG
				pS.Color[CIX_FG] = []int{biCodes[i]}

			} else if isBtween(biCodes[i], 40, 47) || isBtween(biCodes[i], 100, 107) {

				// CLASSIC BG
				pS.Color[CIX_BG] = []int{biCodes[i]}

			} else {

				return fmt.Errorf("SGR-SKIP [UNKWN]: %d", biCodes[i])
			}
		}
	}

	return nil
}

func isBtween(v, lo, hi int) bool {
	return (v >= lo) && (v <= hi)
}

/*
	Formats high color SGR codes
*/
func HighColor(arCodes []int) ([]int, int) {

	nCodes := len(arCodes)

	fnKosher := func(i int) bool {

		return ((i >= 0) && (i <= 255))
	}

	if nCodes >= 3 {

		switch arCodes[1] {

		// 5;n where n is color index (0..255)
		case 5:

			if fnKosher(arCodes[2]) {

				return arCodes[0:3], 2
			}

		// 2;r;g;b where r,g,b are red, green and blue color channels (out of 255)
		case 2:

			if nCodes >= 5 {

				if fnKosher(arCodes[2]) && fnKosher(arCodes[3]) && fnKosher(arCodes[4]) {

					return arCodes[0:5], 4
				}
			}
		}
	}

	return []int{}, 0
}
