/*

	ansiart2utf8.go
	Copyright (C) 2017 Eggplant Systems and Design, LLC

	This program is free software; you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation; either version 2 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License along
	with this program; if not, write to the Free Software Foundation, Inc.,
	51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"

	"ansiart2utf8/ansi"
)

const (
	CHR_ESCAPE = 0x1B
	CHR_CR     = 0x0D
	CHR_LF     = 0x0A
)

func main() {

	var (
		oErr  error
		pFile *os.File
	)

	// ERROR LOGGING
	pLogErr := log.New(os.Stderr, "", log.Lshortfile)
	fnErrExit := func(oErr error) {

		if oErr != nil {

			pLogErr.Output(2, oErr.Error())
			os.Exit(1)
		}
	}

	// TRANSLATION ARRAY
	Array437 := [256]rune{
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

	runtime.GOMAXPROCS(1)
	pFile = nil

	const SZ_HELP_PREFIX = `
ansiart2utf8 VERSION 0.3 BETA
	Converts ANSI art to UTF-8 encoding, expands cursor forward ESC sequences
	into spaces, wraps/resets at a specified line width, sends result to STDOUT.

USAGE: ansiart2utf8 [OPTION]...

OPTIONS
`

	// HELP MESSAGE
	flag.Usage = func() {

		fmt.Fprint(os.Stdout, SZ_HELP_PREFIX)
		flag.PrintDefaults()
		fmt.Fprint(os.Stdout, "\n")
	}

	// COMMAND PARAMETERS
	puiWidth := flag.Uint("w", 80, "LINE WIDTH")
	pszInput := flag.String("f", "-", "INPUT FILENAME, OR \"-\" FOR STDIN")
	pbDebug := flag.Bool("d", false, "DEBUG MODE: LINE NUMBERING + PIPE @ \\n")
	pnRowBytes := flag.Uint("bytes", 0, "MAXIMUM OUTPUT BYTES PER-ROW (0 = NO LIMIT)")

	flag.Parse()

	if !flag.Parsed() {

		fnErrExit(fmt.Errorf("Invalid Parameters"))
	}

	// DEBUG LOGGING
	pLogDebug := log.New(os.Stdout, "", 0)
	fnDebug := func(v ...interface{}) {

		if *pbDebug {
			pLogDebug.Output(2, fmt.Sprint(v...))
		}
	}

	// GET FILE HANDLE
	if strings.Compare(*pszInput, "-") == 0 {

		pFile = os.Stdin

	} else {

		pFile, oErr = os.Open(*pszInput)
		fnErrExit(oErr)

		fnDebug("FILE: ", *pszInput)
	}

	bsInput, oErr := ioutil.ReadAll(pFile)
	fnErrExit(oErr)

	bEsc := false

	bsSGR := ansi.SGR{}
	bsSGR.Reset()
	pGrid := ansi.GridNew(ansi.GridDim(*puiWidth))

	// BUFFER OUTPUT
	pWriter := bufio.NewWriter(os.Stdout)

	curCode := ansi.ECode{}
	curPos := ansi.NewPos()
	curSaved := ansi.NewPos()

	// ITERATE BYTES IN INPUT
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
			if strings.IndexByte(ansi.CodeTerminators(), chr) != -1 {

				bEsc = false
				curCode.Code = rune(chr)

				if curCode.Validate() {

					// ONLY RESTORE SGR ESCAPE CODES
					switch curCode.Code {

					case 'm':

						oErr = bsSGR.Merge(curCode.SubParams)

						if oErr != nil {
							fnDebug(oErr)
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

						curPos.Y = ansi.GridDim(curCode.SubParams[0])
						curPos.X = ansi.GridDim(curCode.SubParams[1])

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

				oErr = pGrid.Put(curPos, ' ', bsSGR)

				if oErr != nil {
					fnDebug(oErr)
				}

				curPos.Y += 1
				curPos.X = 1

			} else if chr != '\b' {

				oErr = pGrid.Put(curPos, Array437[chr], bsSGR)

				if oErr != nil {
					fnDebug(oErr)
				}

				pGrid.Inc(&curPos)
			}
		}
	}

	pGrid.Print(pWriter, int(*pnRowBytes), *pbDebug)

	pWriter.WriteByte(CHR_LF)
	pWriter.Flush()

	os.Exit(0)
}
