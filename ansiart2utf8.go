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
   "flag"
   "fmt"
   "os"
   "runtime"
   "strings"
   "io/ioutil"
   "bufio"
   "regexp"
   "strconv"
)

const (

   CHR_ESCAPE = 0x1B
   CHR_CR     = 0x0D
   CHR_LF     = 0x0A

   // RESET AND CHANGE TO WHITE ON BLACK
   SZ_ESC_RESET = "\x1B[0m\x1B[37;40m"

   // RESET ALL TO DEFAULTS
   SZ_ESC_FINAL_RESET = "\x1B[0m"
)

func fnErrExit(oErr error) {

   fnErrExitEx(oErr, "")
}

func fnErrExitEx(oErr error, szMsg string) {

   if oErr != nil {

      if len(szMsg) > 0 {

         fmt.Fprint(os.Stderr, szMsg, ":\n")
      }

      fmt.Fprint(os.Stderr, oErr, "\n")

      os.Exit(1)
   }
}

func main() {

   var (
      oErr   error    = nil
      pFile  *os.File
   )

   Array437 := [256]rune {
      '\x00','☺','☻','♥','♦','♣','♠','•','\b','\t','\n','♂','♀','\r','♫','☼',
      '►','◄','↕','‼','¶','§','▬','↨','↑','↓','→','\x1b','∟','↔','▲','▼',
      ' ','!','"','#','$','%','&','\'','(',')','*','+',',','-','.','/',
      '0','1','2','3','4','5','6','7','8','9',':',';','<','=','>','?',
      '@','A','B','C','D','E','F','G','H','I','J','K','L','M','N','O',
      'P','Q','R','S','T','U','V','W','X','Y','Z','[','\\',']','^','_',
      '`','a','b','c','d','e','f','g','h','i','j','k','l','m','n','o',
      'p','q','r','s','t','u','v','w','x','y','z','{','|','}','~','⌂',
      '\u0080','\u0081','é','â','ä','à','å','ç','ê','ë','è','ï','î','ì','Ä','Å',
      'É','æ','Æ','ô','ö','ò','û','ù','ÿ','Ö','Ü','¢','£','¥','₧','ƒ',
      'á','í','ó','ú','ñ','Ñ','ª','º','¿','⌐','¬','½','¼','¡','«','»',
      '░','▒','▓','│','┤','╡','╢','╖','╕','╣','║','╗','╝','╜','╛','┐',
      '└','┴','┬','├','─','┼','╞','╟','╚','╔','╩','╦','╠','═','╬','╧',
      '╨','╤','╥','╙','╘','╒','╓','╫','╪','┘','┌','█','▄','▌','▐','▀',
      'α','ß','Γ','π','Σ','σ','µ','τ','Φ','Θ','Ω','δ','∞','φ','ε','∩',
      '≡','±','≥','≤','⌠','⌡','÷','≈','°','∙','·','√','ⁿ','²','■','\u00a0',
   }

   runtime.GOMAXPROCS(1)
   pFile = nil

   const SZ_HELP_PREFIX = `
ansiart2utf8 VERSION 0.1 BETA
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
   puiWidth := flag.Uint(  "w",     80,    "LINE WIDTH")
   pszInput := flag.String("f",     "-",   "INPUT FILENAME, OR \"-\" FOR STDIN")

   flag.Parse()

   if !flag.Parsed() {

      fnErrExit(fmt.Errorf("Invalid Parameters"))
   }

   // GET FILE HANDLE
   if strings.Compare(*pszInput, "-") == 0 {

      pFile = os.Stdin

   } else {

      pFile, oErr = os.Open(*pszInput)
      fnErrExit(oErr)
   }

   bsInput, oErr := ioutil.ReadAll(pFile)

   var (
      bEsc        bool     = false
      lenLine     uint     = 0
      szTempEsc   string   = ""

      bsSGR       []string
   )

   bsSGR = make([]string, 0)

   // BUFFER OUTPUT
   pWriter := bufio.NewWriter(os.Stdout)

   // ESC[nC MOTION
   oREXP, oErr := regexp.Compile("\x1B\\[([[:digit:]]+)C")
   fnErrExit(oErr)

   // VALID RESTORABLE ESC RESET
   oREXP_Reset, oErr := regexp.Compile("\x1B\\[[0]+m")
   fnErrExit(oErr)

   // EXPAND "CURSOR FORWARD" ESC CODES TO ACTUAL SPACES
   szFiltered := oREXP.ReplaceAllStringFunc(string(bsInput), func(match string) string {

      szTmp := oREXP.FindStringSubmatch(match)

      nSpaces, oErr := strconv.ParseUint(szTmp[1], 10, 32)
      fnErrExitEx(oErr, "FAILED TO PARSE MOTION ESCAPE CODE")

      return strings.Repeat(" ", int(nSpaces))
   })

   // ITERATE BYTES IN INPUT
   for _, chr := range []byte(szFiltered) {

      // DROP \r
      if chr == CHR_CR {

         continue

      // BEGIN ESCAPE CODE
      } else if chr == CHR_ESCAPE {

         bEsc = true
         szTempEsc = string(chr)

      // HANDLE ESCAPE CODE SEQUENCE
      } else if bEsc {

         // ESCAPE CODE TERMINATING CHARS:
         // EXIT ESCAPE CODE FSM SUCCESSFULLY ON TERMINATING 'm' CHARACTER
         if strings.IndexByte("mhlJK", chr) != -1 {

            bEsc = false
            szTempEsc += string(chr)

            // ONLY RESTORE SGR ESCAPE CODES
            if chr == 'm' {

               // RESET BGR STACK ON RESET ESCAPE CODE
               if oREXP_Reset.MatchString(szTempEsc) {

                  bsSGR = make([]string, 0)
                  szTempEsc = SZ_ESC_RESET // RESET TO WHITE-ON-BLACK + ^[[0m

               // OTHERWISE, PUSH TO BGR STACK
               } else {

                  bsSGR = append(bsSGR, szTempEsc)
               }
            }

            pWriter.WriteString(szTempEsc)

            continue

         // SKIP + IGNORE CONTROL CHARS DURING ESCAPE CODE
         } else if (chr > 0) && (chr <= 31) {

            continue

         // EXIT ESCAPE CODE FSM ON INVALID CHAR
         } else if strings.IndexByte("0123456789[]noNOPX?^_c;\\", chr) == -1 {

            bEsc = false
            szTempEsc = ""

         // WRITE OUT COMPONENT OF ESCAPE SEQUENCE
         } else {

            szTempEsc += string(chr)
         }
      }

      // HANDLE WRITABLE CHARACTERS OUTSIDE OF ESCAPE MODE
      if !bEsc {

         // WRAP TO NEXT LINE ON \n, OR WHEN SPECIFIED LINE WIDTH IS MET
         if (chr == CHR_LF) || (lenLine == *puiWidth) {

            // RESET LINE, INSERT LF, RESTORE SGR (COLOR/FONT) SETTINGS
            pWriter.WriteString(SZ_ESC_RESET)
            pWriter.WriteByte(CHR_LF)

            for _, szEsc := range bsSGR {

               pWriter.WriteString(szEsc)
            }

            lenLine = 0

            if chr == CHR_LF {

               continue
            }
         }

         lenLine += 1

         // CHARACTER TRANSLATION
         pWriter.WriteRune(Array437[chr])
      }
   }

   pWriter.WriteString(SZ_ESC_FINAL_RESET)
   pWriter.WriteByte(CHR_LF)

   pWriter.Flush()

   os.Exit(0)
}
