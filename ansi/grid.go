package ansi

import (
   "fmt"
   "io"
//   "strconv"
   "strings"
)

const (
   TO_BEGIN = iota
   TO_END   = iota
   ALL      = iota
)

const (
   CLEAR_CHAR   = ' '
   SZ_ESC_RESET = "\x1B[0m\x1B[37;40m"
)

type GridDim uint64

type GridPos struct {

   X, Y  GridDim
}

func NewPos() GridPos {

   return GridPos{ X: 1, Y: 1 }
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

type GridCell struct {

   Char     rune
   Esc      *SGR
}

func (gc *GridCell) Clear() {

   gc.Char = CLEAR_CHAR
   gc.Esc  = nil
}

type GridRow []GridCell

type Grid struct {

   width    GridDim
   grid     []GridRow
}

func GridNew(nWidth GridDim) *Grid {

   sRow  := make([]GridCell, nWidth)
   sGrid := make([]GridRow, 1)
   sGrid[0] = sRow

   return &Grid{
      width:  nWidth,
      grid:   sGrid,
   }
}

func (gr *Grid) Height() GridDim {

   return GridDim(len(gr.grid))
}

type SGR struct {

   Reset, Blink, Bold, Inverse, HighInt, Strikethrough, Underline bool
   Bg_HighInt  bool
   ColorTxt    string
   ColorBg     string
}

func SGR_Reset() SGR {

   oSGR := SGR{}
   oSGR.SGR_Reset()
   return oSGR
}

func (sCodes *SGR) SGR_Reset() {

   if sCodes == nil {
      return
   }

   *sCodes = SGR{
      Reset:    true,
      ColorTxt: "37",
      ColorBg:  "40",
   }
}

func (sCodes *SGR) String() string {

   sRet     := ""

   if sCodes != nil {

      bsParts := []string{}

      if sCodes.Reset {
         bsParts = append(bsParts, "0")
      }

      if sCodes.Bold {
         bsParts = append(bsParts, "1")
      }

      if sCodes.Blink {
         bsParts = append(bsParts, "5")
      }

      if len(sCodes.ColorTxt) > 0 {
         bsParts = append(bsParts, sCodes.ColorTxt)
      }

      if len(sCodes.ColorBg) > 0 {
         bsParts = append(bsParts, sCodes.ColorBg)
      }

      return "\x1B[" + strings.Join(bsParts, ";") + "m"
   }

   return sRet
}

func (pSGR *SGR) Clone() *SGR {

   if pSGR != nil {

      oNew := *pSGR
      return &oNew

   } else {

      return &SGR{}
   }
}

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

/*
   TODO: SOME OF BLOCKTRONICS IS BROKEN
   TODO: COLOR PRESERVATION AFTER MOTION DRAW (transient bit?)
*/

   // TODO: pSGR = nil FOR B&W MODE
   if pSGR == nil {
      return fmt.Errorf("SGR.Merge() called on nil pointer!")
   }

   nCodes := len(biCodes)

   for i := 0; i < nCodes; i++ {

      switch( biCodes[i] ) {

      // RESET
      case 0:

         pSGR.SGR_Reset()
         pSGR.Reset = true

      case 1:

         pSGR.Bold = true

      case 5, 6:

         pSGR.Blink = true

      // DEFAULT FG
      case 39:

         pSGR.ColorTxt = "37"

      // DEFAULT BG
      case 49:

         pSGR.ColorBg = "40"

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

            // TODO: ENABLE/DISABLE W/DEBUG MODE
            return fmt.Errorf("SGR SKIPPED: %d", biCodes[i])
         }
      }
   }

   return nil
}

func (gr *Grid) Print(iWri io.Writer, bDebug bool) {

   for nRow, sRow := range gr.grid {

      if bDebug { fmt.Fprintf(iWri, "%5d: ", nRow + 1) }

      szLastCode := ""
      szCurCode  := ""

      for _, cell := range sRow {

         c := cell.Char

         switch( c ) {

         case 0:
            c = CLEAR_CHAR

         case '\n':
            goto NEXT_LINE
         }

         if cell.Esc == nil {

            c = ' '
         }

         // TODO: NULL Esc HANDLING
         szCurCode = cell.Esc.String()

         if( szCurCode != szLastCode ) {

            fmt.Fprint(iWri, szCurCode)
            szLastCode = szCurCode
         }
         fmt.Fprintf(iWri, "%c", c)
      }

NEXT_LINE:

      fmt.Fprint(iWri, SZ_ESC_RESET)

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

func (gr *Grid) Touch(nLine GridDim) GridRow {

   // ENHEIGHTEN
   if nLine >= gr.Height() {

      oldHeight := gr.Height()
      sGrid := make([]GridRow, nLine + 1)
      copy(sGrid, gr.grid)

      for i := oldHeight; i <= nLine; i++ {

         sGrid[i] = make([]GridCell, gr.width)
      }

      gr.grid = sGrid
   }

   return gr.grid[nLine]
}

func (gr *Grid) Put(pos GridPos, rChar *rune, bsCodes SGR) {

   if (pos.X == 0) || (pos.Y == 0) {
      return
   }

   nLine := GridDim(pos.Y - 1)
   nCol  := GridDim(pos.X - 1)

   row := gr.Touch(nLine)

   // TODO: MODULUS/ALLOCATE HANDLING FOR OUT-OF-BOUNDS COLUMNS?
   if nCol < GridDim(len(row)) {

      if rChar != nil {
         row[nCol].Char = *rChar
      }

      row[nCol].Esc = bsCodes.Clone()
   } else {

      // TODO: REMOVE ON RELEASE
      fmt.Printf("INVALID POS: %+v\n", pos)
   }
}

func (gr *Grid) ClearLine(nMode uint, nLine, nCol GridDim) {

   var i GridDim

   row := gr.Touch(nLine)

   switch( nMode ) {

   case TO_BEGIN:

      for i = 0; (i <= nCol) && (i < gr.width); i++ {
         row[i].Clear()
      }

   case TO_END:

      for i = nCol; i < gr.width; i++ {
         row[i].Clear()
      }

   case ALL:

      for i = 0; i < gr.width; i++ {
         row[i].Clear()
      }
   }
}