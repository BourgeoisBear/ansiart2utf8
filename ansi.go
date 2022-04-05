package ansiart2utf8

import (
	"fmt"
	"strconv"
	"strings"
)

// https://www.gnu.org/software/screen/manual/html_node/Control-Sequences.html
const SGR_TERMINATORS = "cfhlmsuABCDEFGHJKNOPSTX\\]^_"

type EscCode struct {
	Params    string
	Code      rune
	SubParams []int
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
	intPrm := make([]int, 0, len(arPrm))
	for _, szCode := range arPrm {

		if n, e := strconv.Atoi(szCode); (e == nil) && (n >= 0) {
			intPrm = append(intPrm, n)
		} else {
			// -1 PLACEHOLDER FOR BLANK/INVALID PARAMS
			intPrm = append(intPrm, -1)
		}
	}

	if strings.IndexRune("su", pC.Code) != -1 {

		// NO PARAMS
		pC.SubParams = []int{}

		return true

	} else if strings.IndexRune("ABCDEFGST", pC.Code) != -1 {

		// ONE PARAM - MOTION

		// DEFAULT
		if len(intPrm) > 0 && intPrm[0] > 1 {
			pC.SubParams = []int{intPrm[0]}
		} else {
			pC.SubParams = []int{1}
		}

		return true

	} else if strings.IndexRune("JK", pC.Code) != -1 {

		// ONE PARAM - ERASE DISPLAY/LINE

		// DEFAULT
		if len(intPrm) > 0 && IsBtween(intPrm[0], 0, 2) {
			pC.SubParams = []int{intPrm[0]}
		} else {
			pC.SubParams = []int{0}
		}

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
