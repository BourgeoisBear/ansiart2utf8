package ansiart2utf8

import (
	"fmt"
	"strconv"
	"strings"
)

/*

	NOTE: 1-BASED INDEXING

	ESC[#h   set mode
	ESC[#l   reset mode

	Esc[Code;String;...p  set keyboard strings

NO EFFECT IF ALREADY AT "EDGE" OF SCREEN:
	ESC[#A         moves cursor up # lines
	ESC[#B         moves cursor down # lines
	ESC[#C         moves cursor right # spaces
	ESC[#D         moves cursor left # spaces

IGNORED SINCE NOT IN ANSI.SYS:
	ESC[#E         moves cursor to beginning of line #-lines down
	ESC[#F         moves cursor to beginning of line #-lines up
	ESC[#G         moves cursor to column #
	ESC[#S         scroll whole page up by # (default 1) lines.
							New lines are added at the bottom.
	ESC[#T         scroll whole page down by # (default 1) lines.
							New lines are added at the top.

POSITION:
	ESC[n;mH       moves cursor to row n, column n (default 1 for omitteds)
	ESC[n;mf       moves cursor to row n, column n (default 1 for omitteds)

CLEAR:
	ESC[J          clear from cursor to end of screen
	ESC[0J         "
	ESC[1J         clear from cursor to beginning of screen
	ESC[2J         clear screen and home cursor
	ESC[3J         clear screen + scrollback buffer, and home cursor [probly subst w/ 2J]
	ESC[K          clear to end of line
	ESC[0K         "
	ESC[1K         clear to beginning of line
	ESC[2K         clear entire line

SAVE/RESTORE:
	ESC[s          save cursor position for recall later
	ESC[u          Return to saved cursor position

SGR:
	ESC[(params)m

*/

// https://www.gnu.org/software/screen/manual/html_node/Control-Sequences.html
const SGR_TERMINATORS = "cfhlmsuABCDEFGHJKNOPSTX\\]^_"

type EscCode struct {
	Params    string
	Code      rune
	SubParams []int
}

// ORIGINAL COLORS
type OC struct {
	Hex      string
	Xterm256 int
}

type ValidateFunc func(pCode *EscCode) bool

var (
	mapValidate = map[rune]ValidateFunc{
		// IGNORE
		'E': VF_Ignore,
		'F': VF_Ignore,
		'G': VF_Ignore,
		'S': VF_Ignore,
		'T': VF_Ignore,
		// NON-CSI + ZERO-PARAMS
		'N':  VF_NonCSI,
		'O':  VF_NonCSI,
		'P':  VF_NonCSI,
		'\\': VF_NonCSI,
		']':  VF_NonCSI,
		'X':  VF_NonCSI,
		'^':  VF_NonCSI,
		'_':  VF_NonCSI,
		'c':  VF_NonCSI,
		// CSI
		'A': CSI_Params,
		'B': CSI_Params,
		'C': CSI_Params,
		'D': CSI_Params,
		's': CSI_Params,
		'u': CSI_Params,
		'J': CSI_Params,
		'K': CSI_Params,
		'H': CSI_Params,
		'f': CSI_Params,
		// SGR
		'm': VF_SGR,
	}
)

func VF_Ignore(pC *EscCode) bool {

	return false
}

func VF_NonCSI(pC *EscCode) bool {

	pC.Params = strings.TrimSpace(pC.Params)

	return (len(pC.Params) == 0)
}

func CSI_Params(pC *EscCode) bool {

	szParams := strings.TrimSpace(pC.Params)

	// MUST BEGIN WITH '['
	if (len(szParams) == 0) || (szParams[0] != '[') {
		return false
	}

	// CONVERT TO INT LIST
	arPrm := strings.Split(szParams[1:], ";")
	intPrm := make([]int, len(arPrm))
	for ix := range arPrm {
		var e error
		if intPrm[ix], e = strconv.Atoi(arPrm[ix]); e != nil {
			intPrm[ix] = -1
		}
	}

	if strings.IndexRune("su", pC.Code) != -1 {

		// NO PARAMS

		return true

	} else if strings.IndexRune("ABCDEFGST", pC.Code) != -1 {

		// ONE PARAM - MOTION

		// DEFAULT
		nVal := 1
		if len(intPrm) > 0 && intPrm[0] > 1 {
			nVal = intPrm[0]
		}
		pC.SubParams = []int{nVal}
		return true

	} else if strings.IndexRune("JK", pC.Code) != -1 {

		// ONE PARAM - ERASE DISPLAY/LINE

		// DEFAULT
		nVal := 0

		if len(intPrm) > 0 && intPrm[0] > 0 {
			nVal = intPrm[0]
		}
		pC.SubParams = []int{nVal}
		return true

	} else if strings.IndexRune("Hf", pC.Code) != -1 {

		// TWO PARAMS

		// DEFAULTS
		pC.SubParams = []int{1, 1}

		for ix := range intPrm {

			if ix >= len(pC.SubParams) {
				break
			}

			if intPrm[ix] > 1 {
				pC.SubParams[ix] = intPrm[ix]
			}
		}

		return true
	}

	return false
}

/*
	H: extends selection
	on jump:
		- colors only apply to written areas, rest remain W on B
		- motions do not count as 'written'

*/
func VF_SGR(pC *EscCode) bool {

	szParams := strings.TrimSpace(pC.Params)

	// MUST BEGIN WITH '['
	if (len(szParams) == 0) || (szParams[0] != '[') {
		return false
	}

	// HANDLE EMPTY ESC[m
	if len(szParams) == 1 {
		pC.SubParams = []int{0}
		return true
	}

	// SPLIT AT ';'
	sPrm := strings.Split(szParams[1:], ";")

	pC.SubParams = []int{}

	// CONVERT TO INT AND APPEND TO PARAMS
	for _, v := range sPrm {

		nVal, err := strconv.Atoi(v)
		if (err != nil) || (nVal < 0) || (nVal > 255) {
			return false
		}

		pC.SubParams = append(pC.SubParams, nVal)
	}

	// TODO: uncomment
	// pC.SubParams = TranslateColors(pC.SubParams)
	return true
}

func (pC *EscCode) Reset() {

	pC.Params = ""
	pC.Code = 0
	pC.SubParams = []int{}
}

func (pC *EscCode) Debug() string {

	return fmt.Sprintf("ESC%s%c; %+v", pC.Params, pC.Code, pC.SubParams)
}

func (pC *EscCode) Validate() bool {

	pC.SubParams = []int{}

	if fnValidate, bOk := mapValidate[pC.Code]; bOk {

		return fnValidate(pC)
	}

	return false
}

func (pC *EscCode) String() string {

	return fmt.Sprintf("\x1B%s%c", pC.Params, pC.Code)
}
