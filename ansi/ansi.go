package ansi

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

type ECode struct {
	Params    string
	Code      rune
	SubParams []int
}

type ValidateFunc func(pCode *ECode) bool

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

// https://www.gnu.org/software/screen/manual/html_node/Control-Sequences.html
func CodeTerminators() string {

	return "cfhlmsuABCDEFGHJKNOPSTX\\]^_"
}

func VF_Ignore(pCode *ECode) bool {

	return false
}

func VF_NonCSI(pCode *ECode) bool {

	pCode.Params = strings.TrimSpace(pCode.Params)

	return (len(pCode.Params) == 0)
}

func CSI_Params(pCode *ECode) bool {

	var nVal int

	szParams := strings.TrimSpace(pCode.Params)

	// MUST BEGIN WITH '['
	if (len(szParams) == 0) || (szParams[0] != '[') {
		return false
	}

	szParams = szParams[1:]

	// NO PARAMS
	if strings.IndexRune("su", pCode.Code) != -1 {

		if len(szParams) == 0 {
			return true
		}

		// ONE PARAM
	} else if strings.IndexRune("ABCDJK", pCode.Code) != -1 {

		// NOTE: RETURNS 0 ON ERROR
		nVal, _ := strconv.Atoi(szParams)

		if ((pCode.Code == 'J') && (nVal > 3)) ||
			((pCode.Code == 'K') && (nVal > 2)) {

			return false
		}

		if nVal < 1 {
			nVal = 1
		}

		pCode.SubParams = []int{nVal}

		return true

		// TWO PARAMS
	} else if strings.IndexRune("Hf", pCode.Code) != -1 {

		pTmp := strings.Split(szParams, ";")

		if len(pTmp) <= 2 {

			subpTemp := []int{1, 1}

			for k, v := range pTmp {

				nVal, _ = strconv.Atoi(v)

				if nVal > 0 {
					subpTemp[k] = nVal
				}
			}

			pCode.SubParams = subpTemp

			return true
		}
	}

	return false
}

func VF_SGR(pCode *ECode) bool {

	szParams := strings.TrimSpace(pCode.Params)

	// MUST BEGIN WITH '['
	if (len(szParams) == 0) || (szParams[0] != '[') {
		return false
	}

	// HANDLE EMPTY ESC[m
	if len(szParams) == 1 {
		pCode.SubParams = []int{0}
		return true
	}

	// SPLIT AT ';'
	pTmp := strings.Split(szParams[1:], ";")

	subpTemp := []int{}

	// CONVERT TO INT AND APPEND TO PARAMS
	for _, v := range pTmp {

		nVal, err := strconv.Atoi(v)

		if (err != nil) || (nVal < 0) || (nVal > 255) {
			return false
		}

		subpTemp = append(subpTemp, nVal)
	}

	pCode.SubParams = subpTemp

	return true
}

func (pCode *ECode) Reset() {

	pCode.Params = ""
	pCode.Code = 0
	pCode.SubParams = []int{}
}

func (pCode *ECode) Debug() string {

	return fmt.Sprintf("ESC%s%c; %+v", pCode.Params, pCode.Code, pCode.SubParams)
}

func (pCode *ECode) Validate() bool {

	pCode.SubParams = []int{}

	fnValidate, bOk := mapValidate[pCode.Code]

	if bOk {

		return fnValidate(pCode)
	}

	return false
}

func (pCode *ECode) String() string {

	return fmt.Sprintf("\x1B%s%c", pCode.Params, pCode.Code)
}
