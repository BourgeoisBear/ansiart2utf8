package ansi

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	CHR_ESCAPE = 0x1B
	CHR_CR     = 0x0D
	CHR_LF     = 0x0A
)

// TRANSLATION ARRAY
var Array437 [256]rune = [256]rune{
	'\x00', '☺', '☻', '♥', '♦', '♣', '♠', '•', '\b', '\t', '\n', '♂', '♀', '\r', '♫', '☼',
	'►', '◄', '↕', '‼', '¶', '§', '▬', '↨', '↑', '↓', '→', '\x1b', '∟', '↔', '▲', '▼',
	' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?',
	'@', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '[', '\\', ']', '^', '_',
	'`', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '{', '|', '}', '~', '⌂',
	'\u0080', '\u0081', 'é', 'â', 'ä', 'à', 'å', 'ç', 'ê', 'ë', 'è', 'ï', 'î', 'ì', 'Ä', 'Å',
	'É', 'æ', 'Æ', 'ô', 'ö', 'ò', 'û', 'ù', 'ÿ', 'Ö', 'Ü', '¢', '£', '¥', '₧', 'ƒ',
	'á', 'í', 'ó', 'ú', 'ñ', 'Ñ', 'ª', 'º', '¿', '⌐', '¬', '½', '¼', '¡', '«', '»',
	'░', '▒', '▓', '│', '┤', '╡', '╢', '╖', '╕', '╣', '║', '╗', '╝', '╜', '╛', '┐',
	'└', '┴', '┬', '├', '─', '┼', '╞', '╟', '╚', '╔', '╩', '╦', '╠', '═', '╬', '╧',
	'╨', '╤', '╥', '╙', '╘', '╒', '╓', '╫', '╪', '┘', '┌', '█', '▄', '▌', '▐', '▀',
	'α', 'ß', 'Γ', 'π', 'Σ', 'σ', 'µ', 'τ', 'Φ', 'Θ', 'Ω', 'δ', '∞', 'φ', 'ε', '∩',
	'≡', '±', '≥', '≤', '⌠', '⌡', '÷', '≈', '°', '∙', '·', '√', 'ⁿ', '²', '■', '\u00a0',
}

func ToUTF8(iRdr io.Reader, iWri io.Writer, nLineWidth, nMaxBytesPerRow uint, pLogDebug *log.Logger) (E error) {

	fnDebug := func(v ...interface{}) {
		if pLogDebug != nil {
			pLogDebug.Output(2, fmt.Sprint(v...))
		}
	}

	bEsc := false
	bsSGR := SGR{}
	bsSGR.Reset()
	pGrid := GridNew(GridDim(nLineWidth))

	// BUFFER OUTPUT
	pWriter := bufio.NewWriter(os.Stdout)

	curCode := ECode{}
	curPos := NewPos()
	curSaved := NewPos()

	// ITERATE BYTES IN INPUT
	bsInput, E := io.ReadAll(iRdr)
	if E != nil {
		return
	}

	for _, chr := range bsInput {

		// DROP \r
		if chr == CHR_CR {

			continue

			// BEGIN ESCAPE CODE
		} else if chr == CHR_ESCAPE {

			bEsc = true
			curCode.Reset()

			// HANDLE ESCAPE CODE SEQUENCE
		} else if bEsc {

			// NOPS

			/*
				UNHANDLED CODE:   ESC[J; [1]
				UNHANDLED CODE:   ESC[K; [1]
				UNHANDLED CODE:   ESCc;
				INVALID CODE:     ESC[MF;
				INVALID CODE:     ESC[m;
				INVALID CODE:     ESC[P;
				INVALID CODE:     ESC[T;
				INVALID CODE:     ESC[S;
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
			// EXIT ESCAPE CODE FSM SUCCESSFULLY ON TERMINATING 'm' CHARACTER
			if strings.IndexByte(CodeTerminators(), chr) != -1 {

				bEsc = false
				curCode.Code = rune(chr)

				if curCode.Validate() {

					// ONLY RESTORE SGR ESCAPE CODES
					switch curCode.Code {

					case 'm':

						if E = bsSGR.Merge(curCode.SubParams); E != nil {
							fnDebug(E)
						}

					// UP
					case 'A':

						pGrid.IncClamp(&curPos, 0, -int(curCode.SubParams[0]))

					// DOWN
					case 'B':

						pGrid.IncClamp(&curPos, 0, int(curCode.SubParams[0]))

					// FORWARD
					case 'C':

						pGrid.IncClamp(&curPos, int(curCode.SubParams[0]), 0)

					// BACK
					case 'D':

						pGrid.IncClamp(&curPos, -int(curCode.SubParams[0]), 0)

					// TO X,Y
					case 'H', 'f':

						curPos.Y = GridDim(curCode.SubParams[0])
						curPos.X = GridDim(curCode.SubParams[1])

					// SAVE CURSOR POS
					case 's':

						curSaved = curPos

					// RESTORE CURSOR POS
					case 'u':

						curPos = curSaved

					default:

						fnDebug("UNHANDLED CODE: ", curCode.Debug())
						continue
					}

					// fnDebug("SUCCESS: ", curCode.Debug())

				} else {

					fnDebug("INVALID CODE: ", curCode.Debug())
				}

				continue

				// SKIP + IGNORE CONTROL CHARS DURING ESCAPE CODE
			} else if (chr > 0) && (chr <= 31) {

				continue

				// WRITE OUT COMPONENT OF ESCAPE SEQUENCE
			} else {

				curCode.Params += string(chr)
			}
		}

		// HANDLE WRITABLE CHARACTERS OUTSIDE OF ESCAPE MODE
		if !bEsc {

			if chr == CHR_LF {

				if E = pGrid.Put(curPos, ' ', bsSGR); E != nil {
					fnDebug(E)
				}

				curPos.Y += 1
				curPos.X = 1

			} else if chr != '\b' {

				if E = pGrid.Put(curPos, Array437[chr], bsSGR); E != nil {
					fnDebug(E)
				}

				pGrid.Inc(&curPos)
			}
		}
	}

	pGrid.Print(pWriter, int(nMaxBytesPerRow), pLogDebug != nil)

	pWriter.WriteByte(CHR_LF)
	pWriter.Flush()
	return
}
