package ansiart2utf8

import (
	"errors"
	"fmt"
	"io"
)

type GridPos struct {
	X, Y int
}

func NewPos() GridPos {
	return GridPos{X: 1, Y: 1}
}

func (pos *GridPos) Denorm() (X, Y int) {

	X = int(pos.X) - 1
	Y = int(pos.Y) - 1

	if X < 0 {
		X = 0
	}

	if Y < 0 {
		Y = 0
	}

	return
}

type GridCell struct {
	Char  rune
	Brush SGR
}

func (gc *GridCell) ClearCell() {
	gc.Char = 0
	gc.Brush.Reset()
}

type GridRow []GridCell

type Grid struct {
	width uint
	grid  []GridRow
}

func NewGrid(nWidth uint) (G Grid, E error) {

	if nWidth < 1 {
		E = errors.New("GRID WIDTH MUST BE > 0")
		return
	}

	G.width = nWidth
	G.Touch(1)
	return
}

func (gr *Grid) Height() int {
	return int(len(gr.grid))
}

func (gr *Grid) Inc(pos *GridPos, nAmt int) {

	if pos == nil {
		return
	}

	defer func() {
		gr.Touch(pos.Y)
	}()

	iW := int(gr.width)
	x, y := pos.Denorm()

	A := (y * iW) + x
	A += nAmt

	if A > 0 {
		pos.X = (A % iW) + 1
		pos.Y = (A / iW) + 1
	} else {
		pos.X, pos.Y = 1, 1
	}
}

/*
	Increment GridPos `pos` by X, Y
	Clamp result to dimensions of Grid `gr`
*/
func (gr *Grid) IncClamp(pos *GridPos, X, Y int) {

	if pos != nil {

		nWid := int(gr.width)
		if nWid < 1 {
			nWid = 1
		}

		nHgt := int(gr.Height())
		if nHgt < 1 {
			nHgt = 1
		}

		X = int(pos.X) + X

		if X < 1 {
			pos.X = int(1)
		} else if X > nWid {
			pos.X = int(nWid)
		} else {
			pos.X = int(X)
		}

		Y = int(pos.Y) + Y

		if Y < 1 {
			pos.Y = int(1)
		} else if Y > nHgt {
			pos.Y = int(nHgt)
		} else {
			pos.Y = int(Y)
		}
	}
}

func (gr *Grid) ClearFromPosToBegin(pos GridPos) {

	colStart, rowStart := pos.Denorm()

	for ixRow, sRow := range gr.grid {

		if ixRow > rowStart {
			break
		}

		for ixCol := range sRow {

			if (ixRow == rowStart) && (ixCol > colStart) {
				break
			}

			sRow[ixCol].ClearCell()
		}
	}
}

func (gr *Grid) ClearFromPosToEnd(pos GridPos) {

	colStart, rowStart := pos.Denorm()

	for ixRow, sRow := range gr.grid {

		if ixRow < rowStart {
			continue
		}

		for ixCol := range sRow {

			if (ixRow == rowStart) && (ixCol < colStart) {
				continue
			}

			sRow[ixCol].ClearCell()
		}
	}
}

func (gr *Grid) ClearLine(pos GridPos, bToBegin bool) {

	colStart, ixRow := pos.Denorm()

	if ixRow >= len(gr.grid) {
		return
	}

	sRow := gr.grid[ixRow]
	for ix := range sRow {

		if bToBegin {
			if ix > colStart {
				break
			}
		} else {
			if ix < colStart {
				continue
			}
		}

		sRow[ix].ClearCell()
	}
}

func (gr *Grid) ResetChars(rChar rune) {
	for _, sRow := range gr.grid {
		for ixCol, _ := range sRow {
			sRow[ixCol].Char = rChar
		}
	}
}

/*
	Extends grid height to `nHeight` if grid is shorter.
*/
func (gr *Grid) Touch(nHeight int) {

	oldHeight := gr.Height()

	if nHeight > oldHeight {

		// ENHEIGHTEN
		sGrid := make([]GridRow, nHeight)
		if oldHeight > 0 {
			copy(sGrid, gr.grid)
		}

		// ADD NEW ROWS
		for i := oldHeight; i < nHeight; i++ {
			sGrid[i] = make([]GridCell, gr.width)
		}

		gr.grid = sGrid
	}
}

func (gr *Grid) Put(pos GridPos, rChar rune, sgrCodes SGR) error {

	// CONVERT TO 1-BASED TO 0-BASED
	if (pos.X == 0) || (pos.Y == 0) {
		return fmt.Errorf("BAD POSITION %d, %d", pos.X, pos.Y)
	}

	ixCol, ixLine := pos.Denorm()
	if gr.width > 0 {

		if ixCol >= int(gr.width) {
			return fmt.Errorf("POSITION %d, %d EXCEEDS COLUMN WIDTH %d", pos.X, pos.Y, gr.width)
		}
	}

	// ALLOCATE GRID UP TO CURRENT POSITION
	gr.Touch(pos.Y)

	// SPACES AS BLANKS
	if rChar == ' ' {
		rChar = 0
	}

	// WRITE CHAR/FORMATTING TO GRID
	row := gr.grid[ixLine]
	row[ixCol].Char = rChar
	row[ixCol].Brush = sgrCodes

	return nil
}

func (gr *Grid) Print(iWri io.Writer, nRowBytes int, bDebug bool) {

	/*
		NOTE: CAN'T ESC[nC COMPRESS BECAUSE OF TERMINAL BACKGROUND COLOR
	*/

	const STR_CLEAR = "\x1b[0m"

	var nBytes int
	var brushPrev SGR

	fnWrite := func(str string) bool {

		nBytes += len(str)

		// LINE-LENGTH LIMITATION
		if (nRowBytes > 0) && (nBytes > nRowBytes) {
			return true
		}

		fmt.Fprint(iWri, str)
		return false
	}

	for nRow, sRow := range gr.grid {

		// RESET PER-LINE BYTE COUNT
		nBytes = 0
		fnWrite(STR_CLEAR)

		if bDebug {
			lineNum := fmt.Sprintf("%5d: ", nRow+1)
			fnWrite(lineNum)
		}

		// TODO: differential encoding of SGRs
		// TODO: optional toggle of ESC reporting
		// RESTORE BRUSH FROM PREVIOUS LINE
		bFakeEsc := false
		if szTemp := brushPrev.ToEsc(bFakeEsc); fnWrite(szTemp) {
			break
		}

		for _, cell := range sRow {

			// TODO: start with B&W for empty color
			// TODO: IsEqual change for blank fg/bg

			// WRITE ESC CODE ON CHANGE
			if !cell.Brush.IsEqual(&brushPrev) {

				if szTemp := cell.Brush.ToEsc(bFakeEsc); fnWrite(szTemp) {
					break
				}

				// TODO: color continuation into `brushPrev` for empty colors
				// i.e. merge into
				brushPrev = cell.Brush
			}

			// DEFAULT PAINT CHAR
			if cell.Char == 0 {
				cell.Char = ' '
			}

			if fnWrite(string(cell.Char)) {
				break
			}
		}

		fnWrite(STR_CLEAR)

		if bDebug {
			fnWrite("|")
		}

		fnWrite("\n")
	}
}
