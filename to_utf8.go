package ansiart2utf8

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// TRANSLATION ARRAY
var Array437 [256]rune = [256]rune{
	'\x00', '☺', '☻', '♥', '♦', '♣', '♠', '•', '◘', '○', '◙', '♂', '♀', '♪', '♫', '☼',
	'►', '◄', '↕', '‼', '¶', '§', '▬', '↨', '↑', '↓', '→', '←', '∟', '↔', '▲', '▼',
	' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?',
	'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_',
	'`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', '⌂',
	'Ç', 'ü', 'é', 'â', 'ä', 'à', 'å', 'ç', 'ê', 'ë', 'è', 'ï', 'î', 'ì', 'Ä', 'Å',
	'É', 'æ', 'Æ', 'ô', 'ö', 'ò', 'û', 'ù', 'ÿ', 'Ö', 'Ü', '¢', '£', '¥', '₧', 'ƒ',
	'á', 'í', 'ó', 'ú', 'ñ', 'Ñ', 'ª', 'º', '¿', '⌐', '¬', '½', '¼', '¡', '«', '»',
	'░', '▒', '▓', '│', '┤', '╡', '╢', '╖', '╕', '╣', '║', '╗', '╝', '╜', '╛', '┐',
	'└', '┴', '┬', '├', '─', '┼', '╞', '╟', '╚', '╔', '╩', '╦', '╠', '═', '╬', '╧',
	'╨', '╤', '╥', '╙', '╘', '╒', '╓', '╫', '╪', '┘', '┌', '█', '▄', '▌', '▐', '▀',
	'α', 'ß', 'Γ', 'π', 'Σ', 'σ', 'µ', 'τ', 'Φ', 'Θ', 'Ω', 'δ', '∞', 'φ', 'ε', '∩',
	'≡', '±', '≥', '≤', '⌠', '⌡', '÷', '≈', '°', '∙', '·', '√', 'ⁿ', '²', '■', '\u00a0',
}

const (
	CHR_ESCAPE = 0x1B
	CHR_CR     = 0x0D
	CHR_LF     = 0x0A
)

type DebugFunc func(...interface{}) (int, error)

type UTF8Marshaller struct {
	Width    uint
	MaxBytes uint
	Xfrm4bit bool
	FakeEsc  bool
	Debug    DebugFunc
	Writer   io.Writer
}

/*
	ENCODES ANSI ART TO MODERN UTF8 TERMINAL CHARS
	PRE-RENDERS TO MEMORY (MOTION ESCAPES, COLOR CHANGES, ETC)
	WRITES OUTPUT, LINE-BY-LINE, TO .Writer
*/
func (M UTF8Marshaller) Encode(rdAnsi io.Reader) (E error) {

	ixByte := -1
	defer func() {

		if E != nil {
			E = fmt.Errorf("%s, at index %d", E.Error(), ixByte)
		}
	}()

	pRdr := bufio.NewReader(rdAnsi)
	pGrid, E := NewGrid(M.Width)
	if E != nil {
		return
	}

	bEsc := false
	escCur := EscCode{}

	sgrCur := SGR{}
	sgrSaved := SGR{}
	posCur := NewPos()
	posSaved := NewPos()

	// NO-OP
	fnDebug := func(v ...interface{}) (int, error) {

		if M.Debug != nil {
			v = append(v, fmt.Sprintf("at index %d", ixByte))
			return M.Debug(v...)
		}

		return 0, nil
	}

CharLoop:

	for true {

		// TODO: read rune at a time
		chr, e := pRdr.ReadByte()

		if e == io.EOF {
			break
		} else if e != nil {
			E = e
			return
		}
		ixByte += 1

		// TODO: verify continuation of SGR code for unpainted chars / across newlines

		switch chr {

		// TODO: break at ^ZSAUCE00 (^Z is 26 dec, 0x1A hex)
		// STOP ON NULL & SAUCE
		case 0, 26:
			break CharLoop
		}

		if chr == CHR_CR {

			posCur.X = 1
			continue

		} else if chr == CHR_LF {

			// EXTEND ROW
			posCur.Y += 1
			pGrid.Touch(posCur.Y)
			continue

		} else if chr == CHR_ESCAPE {

			// BEGIN ESCAPE CODE
			bEsc = true
			escCur.Reset()
			continue

			// HANDLE ESCAPE CODE SEQUENCE
		} else if bEsc {

			// NOPS

			/*
				UNHANDLED CODE:   ESCc;
				INVALID CODE:     ESC[MF;
				INVALID CODE:     ESC[m;
				INVALID CODE:     ESC[P;
				UNHANDLED CODE:   ESC[@K; [1]
				INVALID CODE:     ESC[@l;
				INVALID CODE:     ESC[@S;
				INVALID CODE:     ESC[@N;
				INVALID CODE:     ESC[@u;
				INVALID CODE:     ESC[@s;
				INVALID CODE:     ESC[Mo3egc;
			*/
			// TODO: ESC[?7h; - possibly "wrap" mode

			// ESCAPE CODE TERMINATING CHARS:
			if strings.IndexByte(SGR_TERMINATORS, chr) == -1 {

				// APPEND COMPONENT OF ESCAPE SEQUENCE
				// TODO: filter non-alphanumerics, & appropriate punctuation
				escCur.Params += string(chr)

			} else {

				// EXIT ESCAPE CODE FSM SUCCESSFULLY ON TERMINATING 'm' CHARACTER

				bEsc = false
				escCur.Code = rune(chr)
				// fmt.Println(escCur.Params + string(escCur.Code))

				if escCur.Validate() {

					// ONLY RESTORE SGR ESCAPE CODES
					switch escCur.Code {

					case 'm':

						if e2 := sgrCur.MergeCodes(escCur.SubParams); e2 != nil {
							E = fmt.Errorf("SGR ERROR %s", e2.Error())
							return
						}

					// UP
					case 'A':

						pGrid.IncClamp(&posCur, 0, -int(escCur.SubParams[0]))

					// DOWN
					case 'B':

						pGrid.IncClamp(&posCur, 0, int(escCur.SubParams[0]))

					// FORWARD
					case 'C':

						pGrid.IncClamp(&posCur, int(escCur.SubParams[0]), 0)
						// pGrid.Touch(posCur.Y)

					// BACK
					case 'D':

						pGrid.IncClamp(&posCur, -int(escCur.SubParams[0]), 0)

					// NOTE: NOT ANSI.SYS
					case 'E', 'F', 'G':
						// TODO:
						// E: beginning on line, n lines down
						// F: beginning on line, n lines up
						// G: cursor to column n

					// TO X,Y
					case 'H', 'f':

						// TODO: verify distinction between CUP [H] & HVP [f]
						posCur.Y = int(escCur.SubParams[0])
						posCur.X = int(escCur.SubParams[1])
						pGrid.Touch(posCur.Y)

					case 'J':

						switch escCur.SubParams[0] {

						// clear from cursor to end of screen
						case 0:
							pGrid.ClearFromPosToEnd(posCur)

						// clear from cursor to beginning of screen
						case 1:
							pGrid.ClearFromPosToBegin(posCur)

						// clear entire screen, move cursor to upper-left
						case 2:
							posCur.X, posCur.Y = 1, 1
							pGrid.ClearFromPosToEnd(posCur)

						// clear entire screen, reset scrollback buffer
						case 3:
							pGrid.ClearFromPosToEnd(GridPos{1, 1})
						}

					case 'K':

						switch escCur.SubParams[0] {

						// clear from cursor to end of line
						case 0:
							pGrid.ClearLine(posCur, false)

						// clear from cursor to beginning of line
						case 1:
							pGrid.ClearLine(posCur, true)

						// clear entire line
						case 2:
							pGrid.ClearLine(GridPos{X: 1, Y: posCur.Y}, false)
						}

					// NO-OP: NOT ANSI.SYS
					case 'S', 'T':
						// TODO:
						// S: scroll page up by n lines
						// T: scroll page down by n lines

					// SAVE CURSOR POS & SGR
					case 's':

						posSaved = posCur
						sgrSaved = sgrCur

					// RESTORE CURSOR POS & SGR
					case 'u':

						posCur = posSaved
						sgrCur = sgrSaved

					default:

						E = fmt.Errorf("UNHANDLED CODE %s", escCur.Debug())
						return
					}

				} else {

					fnDebug("INVALID CODE: ", escCur.Debug())
				}

				continue
			}
		}

		// HANDLE WRITABLE CHARACTERS OUTSIDE OF ESCAPE MODE
		if !bEsc {

			if e2 := pGrid.Put(posCur, Array437[chr], sgrCur); e2 != nil {
				fnDebug(e2)
			}

			pGrid.Inc(&posCur, 1)
		}
	}

	pGrid.Print(M.Writer, int(M.MaxBytes), M.Debug != nil, M.FakeEsc)
	return
}
