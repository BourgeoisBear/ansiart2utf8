package ansi

import (
   "fmt"
   "io"
   "strings"
)

const (
   TO_BEGIN = iota
   TO_END   = iota
   ALL      = iota
)

const (
   COLOR_DEFAULT_TXT = "37"
   COLOR_DEFAULT_BG  = "40"
   CLEAR_CHAR        = ' '
)

type GridDim uint64

type GridPos struct {

   X, Y  GridDim
}

func NewPos() GridPos {

   return GridPos{ X: 1, Y: 1 }
}

type SGR struct {

   Bold, Faint, Italic, Underline, Blink, Inverse, Conceal, Strikethrough bool
   ColorTxt    string
   ColorBg     string
}

func (sCodes *SGR) Reset() {

   if sCodes == nil {
      return
   }

   *sCodes = SGR{
      ColorTxt:   COLOR_DEFAULT_TXT,
      ColorBg:    COLOR_DEFAULT_BG,
   }
}

type GridCell struct {

   Char     rune
   Brush    SGR
}

func (gc *GridCell) ClearCell() {

   gc.Char = CLEAR_CHAR
   gc.Brush.Reset()
}

type GridRow []GridCell

type Grid struct {

   width    GridDim
   grid     []GridRow
}

func GridNew(nWidth GridDim) *Grid {

   return &Grid{
      width:  nWidth,
      grid:   make([]GridRow, 0),
   }
}

func (gr *Grid) touch(nRow GridDim) GridRow {

   if nRow >= gr.Height() {

      // ENHEIGHTEN
      oldHeight := gr.Height()
      sGrid := make([]GridRow, nRow + 1)
      copy(sGrid, gr.grid)

      // ADD NEW ROWS
      for i := oldHeight; i <= nRow; i++ {

         sGrid[i] = make([]GridCell, gr.width)

         for j, _ := range sGrid[i] {

            sGrid[i][j].Brush.Reset()
         }
      }

      gr.grid = sGrid
   }

   return gr.grid[nRow]
}

func (gr *Grid) Height() GridDim {

   return GridDim(len(gr.grid))
}

func (gr *Grid) Inc(pos *GridPos) {

   if pos != nil {

      if pos.X < gr.width {

         pos.X += 1

      } else {

         pos.Y += 1
         pos.X = 1
      }
   }
}

func (gr *Grid) IncClamp(pos *GridPos, X, Y int) {

   if pos != nil {

      X = int(pos.X) + X

      if (X > 0) && (X <= int(gr.width)) {

         pos.X = GridDim(X)
      }

      Y = int(pos.Y) + Y

      if (Y > 0) && (Y <= int(gr.Height())) {

         pos.Y = GridDim(Y)
      }
   }
}

// TODO: RETURN ERROR
func HighColor(arCodes []int) (string, int) {

   nCodes := len(arCodes)

   fnKosher := func(i int) bool {

      return ((i >= 0) && (i <= 255))
   }

   if nCodes >= 3 {

      switch( arCodes[1] ) {

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

func (pSGR *SGR) Merge(biCodes []int) error {

   // TODO: pSGR = nil FOR B&W MODE
   if pSGR == nil {
      return fmt.Errorf("SGR.Merge() called on nil pointer!")
   }

   nCodes := len(biCodes)

   for i := 0; i < nCodes; i++ {

      switch( biCodes[i] ) {

      // RESET
      case 0:
         pSGR.Reset()

      // BOLD
      case 1:
         pSGR.Bold = true
      case 21:
         pSGR.Bold = false

      // FAINT
      case 2:
         pSGR.Faint = true
      case 22:
         pSGR.Faint = false

      // ITALIC
      case 3:
         pSGR.Italic = true
      case 23:
         pSGR.Italic = false

      // UNDERLINE
      case 4:
         pSGR.Underline = true
      case 24:
         pSGR.Underline = false

      // BLINK
      case 5, 6:
         pSGR.Blink = true
      case 25:
         pSGR.Blink = false

      // INVERSE
      case 7:
         pSGR.Inverse = true
      case 27:
         pSGR.Inverse = false

      // CONCEAL
      case 8:
         pSGR.Conceal = true
      case 28:
         pSGR.Conceal = false

      // STRIKETHROUGH
      case 9:
         pSGR.Strikethrough = true
      case 29:
         pSGR.Strikethrough = false

      // DEFAULT FG
      case 39:
         pSGR.ColorTxt = COLOR_DEFAULT_TXT

      // DEFAULT BG
      case 49:
         pSGR.ColorBg = COLOR_DEFAULT_BG

      // HIGH COLOR FG
      case 38:

         if szColor, nAdvance := HighColor(biCodes[i:]); len(szColor) > 0 {
            pSGR.ColorTxt = szColor
            i += nAdvance
            continue
         }

      // HIGH COLOR BG
      case 48:

         if szColor, nAdvance := HighColor(biCodes[i:]); len(szColor) > 0 {
            pSGR.ColorBg = szColor
            i += nAdvance
            continue
         }

      default:

         // CLASSIC FG
         if (biCodes[i] >= 30) && (biCodes[i] <= 37) {

            pSGR.ColorTxt = fmt.Sprintf("%d", biCodes[i])

         // CLASSIC BG
         } else if (biCodes[i] >= 40) && (biCodes[i] <= 47) {

            pSGR.ColorBg = fmt.Sprintf("%d", biCodes[i])

         } else {

            return fmt.Errorf("SGR SKIPPED: %d", biCodes[i])
         }
      }
   }

   return nil
}

func (sCodes *SGR) IsEqual(pCodePrev *SGR) bool {

   return (
      (sCodes.Bold            == pCodePrev.Bold) &&
      (sCodes.Faint           == pCodePrev.Faint) &&
      (sCodes.Italic          == pCodePrev.Italic) &&
      (sCodes.Underline       == pCodePrev.Underline) &&
      (sCodes.Blink           == pCodePrev.Blink) &&
      (sCodes.Inverse         == pCodePrev.Inverse) &&
      (sCodes.Conceal         == pCodePrev.Conceal) &&
      (sCodes.Strikethrough   == pCodePrev.Strikethrough) &&
      (sCodes.ColorTxt        == pCodePrev.ColorTxt) &&
      (sCodes.ColorBg         == pCodePrev.ColorBg))
}

func (sCodes *SGR) ToEsc(pCodePrev *SGR) string {

   if (sCodes != nil) && (pCodePrev != nil) {

      bsParts := []string{}

      bsIter := []struct{BCurrent bool; Set string; Clear string; BPrev bool}{
         {sCodes.Bold,          "1", "21", pCodePrev.Bold},
         {sCodes.Faint,         "2", "22", pCodePrev.Faint},
         {sCodes.Italic,        "3", "23", pCodePrev.Italic},
         {sCodes.Underline,     "4", "24", pCodePrev.Underline},
         {sCodes.Blink,         "5", "25", pCodePrev.Blink},
         {sCodes.Inverse,       "7", "27", pCodePrev.Inverse},
         {sCodes.Conceal,       "8", "28", pCodePrev.Conceal},
         {sCodes.Strikethrough, "9", "29", pCodePrev.Strikethrough},
      }

      // 1/21  X  bold
      // 2/22  X  faint, normal intensity
      // 3/23  X  italic
      // 4/24  X  underline
      // 5/6   X  blink, 25 blink-off
      // 7/27  X  inverse
      // 8/28  X  conceal/reveal
      // 9/29  X  strikethrough

      for _, sITER := range bsIter {

         if sITER.BCurrent && !sITER.BPrev {
            bsParts = append(bsParts, sITER.Set)
         } else if !sITER.BCurrent && sITER.BPrev {
            bsParts = append(bsParts, sITER.Clear)
         }
      }

      if (len(sCodes.ColorTxt) > 0) && (sCodes.ColorTxt != pCodePrev.ColorTxt) {
         bsParts = append(bsParts, sCodes.ColorTxt)
      }

      if (len(sCodes.ColorBg) > 0) && (sCodes.ColorBg != pCodePrev.ColorBg) {
         bsParts = append(bsParts, sCodes.ColorBg)
      }

      if len(bsParts) > 0 {
         return "\x1B[" + strings.Join(bsParts, ";") + "m"
      }
   }

   return ""
}

func (gr *Grid) Print(iWri io.Writer, bDebug bool) {

/*
   NOTE: CAN'T ESC[nC COMPRESS BECAUSE OF TERMINAL BACKGROUND COLOR
      - MAKE THIS AN OPTION FOR RUNS OF SPACE W/BG COLOR?
*/

   // RESET BRUSH
   fmt.Fprint(iWri, "\x1B[0m")

   for nRow, sRow := range gr.grid {

      if bDebug { fmt.Fprintf(iWri, "%5d: ", nRow + 1) }

      brushPrev := SGR{}

      for _, cell := range sRow {

         if cell.Char == '\n' {
            goto NEXT_LINE
         }

         // WRITE ESC CODE ON CHANGE
         if !cell.Brush.IsEqual(&brushPrev) {

            szCode := cell.Brush.ToEsc(&brushPrev)
            fmt.Fprint(iWri, szCode)
            brushPrev = cell.Brush
         }

         if cell.Char == 0 {
            cell.Char = CLEAR_CHAR
         }

         fmt.Fprintf(iWri, "%c", cell.Char)
      }

NEXT_LINE:

      fmt.Fprint(iWri, "\x1B[0m")

      if bDebug { fmt.Fprint(iWri, "|") }

      fmt.Fprint(iWri, "\n")
   }
}

func (gr *Grid) ResetChars(rChar rune) {

   for _, sRow := range gr.grid {

      for ixCol, _ := range sRow {

         sRow[ixCol].Char = rChar
      }
   }
}

func (pos GridPos) ErrInvalid() error {

   return fmt.Errorf("INVALID POSITION %d, %d", pos.X, pos.Y)
}

func (pos GridPos) Normalize() (GridDim, GridDim, error) {

   if (pos.X == 0) || (pos.Y == 0) {
      return 0, 0, pos.ErrInvalid()
   }

   return GridDim(pos.X - 1), GridDim(pos.Y - 1), nil
}

func (gr *Grid) Put(pos GridPos, rChar *rune, sgrCodes SGR) error {

   nCol, nLine, oErr := pos.Normalize()

   if oErr != nil { return oErr }

   row := gr.touch(nLine)

   // TODO: MODULUS/ALLOCATE HANDLING FOR OUT-OF-BOUNDS COLUMNS?
   if nCol < GridDim(len(row)) {

      if rChar != nil {
         row[nCol].Char = *rChar
      }

      row[nCol].Brush = sgrCodes

   } else {

      return pos.ErrInvalid()
   }

   return nil
}

func (gr *Grid) ClearLine(pos GridPos, nMode uint) error {

   var i GridDim

   nCol, nLine, oErr := pos.Normalize()

   if oErr != nil { return oErr }

   row := gr.touch(nLine)

   switch( nMode ) {

   case TO_BEGIN:

      for i = 0; (i <= nCol) && (i < gr.width); i++ {
         row[i].ClearCell()
      }

   case TO_END:

      for i = nCol; i < gr.width; i++ {
         row[i].ClearCell()
      }

   case ALL:

      for i = 0; i < gr.width; i++ {
         row[i].ClearCell()
      }
   }

   return nil
}