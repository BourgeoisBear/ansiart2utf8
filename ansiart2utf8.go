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

var (
   Map437 map[byte]rune
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

   Map437 = map[byte]rune {
      1:    '\u263A',
      2:    '\u263B',
      3:    '\u2665',
      4:    '\u2666',
      5:    '\u2663',
      6:    '\u2660',
      7:    '\u2022',
      11:   '\u2642',
      12:   '\u2640',
      14:   '\u266B',
      15:   '\u263C',
      16:   '\u25BA',
      17:   '\u25C4',
      18:   '\u2195',
      19:   '\u203C',
      20:   '\u00B6',
      21:   '\u00A7',
      22:   '\u25AC',
      23:   '\u21A8',
      24:   '\u2191',
      25:   '\u2193',
      26:   '\u2192',
      28:   '\u221F',
      29:   '\u2194',
      30:   '\u25B2',
      31:   '\u25BC',
      127:  '\u2302',
      130:  '\u00E9',
      131:  '\u00E2',
      132:  '\u00E4',
      133:  '\u00E0',
      134:  '\u00E5',
      135:  '\u00E7',
      136:  '\u00EA',
      137:  '\u00EB',
      138:  '\u00E8',
      139:  '\u00EF',
      140:  '\u00EE',
      141:  '\u00EC',
      142:  '\u00C4',
      143:  '\u00C5',
      144:  '\u00C9',
      145:  '\u00E6',
      146:  '\u00C6',
      147:  '\u00F4',
      148:  '\u00F6',
      149:  '\u00F2',
      150:  '\u00FB',
      151:  '\u00F9',
      152:  '\u00FF',
      153:  '\u00D6',
      154:  '\u00DC',
      155:  '\u00A2',
      156:  '\u00A3',
      157:  '\u00A5',
      158:  '\u20A7',
      159:  '\u0192',
      160:  '\u00E1',
      161:  '\u00ED',
      162:  '\u00F3',
      163:  '\u00FA',
      164:  '\u00F1',
      165:  '\u00D1',
      166:  '\u00AA',
      167:  '\u00BA',
      168:  '\u00BF',
      169:  '\u2310',
      170:  '\u00AC',
      171:  '\u00BD',
      172:  '\u00BC',
      173:  '\u00A1',
      174:  '\u00AB',
      175:  '\u00BB',
      176:  '\u2591',
      177:  '\u2592',
      178:  '\u2593',
      179:  '\u2502',
      180:  '\u2524',
      181:  '\u2561',
      182:  '\u2562',
      183:  '\u2556',
      184:  '\u2555',
      185:  '\u2563',
      186:  '\u2551',
      187:  '\u2557',
      188:  '\u255D',
      189:  '\u255C',
      190:  '\u255B',
      191:  '\u2510',
      192:  '\u2514',
      193:  '\u2534',
      194:  '\u252C',
      195:  '\u251C',
      196:  '\u2500',
      197:  '\u253C',
      198:  '\u255E',
      199:  '\u255F',
      200:  '\u255A',
      201:  '\u2554',
      202:  '\u2569',
      203:  '\u2566',
      204:  '\u2560',
      205:  '\u2550',
      206:  '\u256C',
      207:  '\u2567',
      208:  '\u2568',
      209:  '\u2564',
      210:  '\u2565',
      211:  '\u2559',
      212:  '\u2558',
      213:  '\u2552',
      214:  '\u2553',
      215:  '\u256B',
      216:  '\u256A',
      217:  '\u2518',
      218:  '\u250C',
      219:  '\u2588',
      220:  '\u2584',
      221:  '\u258C',
      222:  '\u2590',
      223:  '\u2580',
      224:  '\u03B1',
      225:  '\u00DF',
      226:  '\u0393',
      227:  '\u03C0',
      228:  '\u03A3',
      229:  '\u03C3',
      230:  '\u00B5',
      231:  '\u03C4',
      232:  '\u03A6',
      233:  '\u0398',
      234:  '\u03A9',
      235:  '\u03B4',
      236:  '\u221E',
      237:  '\u03C6',
      238:  '\u03B5',
      239:  '\u2229',
      240:  '\u2261',
      241:  '\u00B1',
      242:  '\u2265',
      243:  '\u2264',
      244:  '\u2320',
      245:  '\u2321',
      246:  '\u00F7',
      247:  '\u2248',
      248:  '\u00B0',
      249:  '\u2219',
      250:  '\u00B7',
      251:  '\u221A',
      252:  '\u207F',
      253:  '\u00B2',
      254:  '\u25A0',
      255:  '\u00A0',
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

   // VALID RESTORABLE ESC
   oREXP_Restorable, oErr := regexp.Compile("\x1B\\[[[:digit:];]+m")
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

         // EXIT ESCAPE CODE FSM SUCCESSFULLY ON TERMINATING 'm' CHARACTER
         if chr == 'm' {

            bEsc = false
            szTempEsc += string(chr)

            // ONLY RESTORE SGR ESCAPE CODES
            if oREXP_Restorable.MatchString(szTempEsc) {

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
         } else if strings.IndexByte("0123456789[]noNOPX^_c;\\", chr) == -1 {

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

         // ATTEMPT CHARACTER TRANSLATION
         strRune, bOk := Map437[chr]

         if( bOk ) {

            pWriter.WriteRune(strRune)

         } else {

            pWriter.WriteByte(chr)
         }
      }
   }

   pWriter.WriteString(SZ_ESC_FINAL_RESET)
   pWriter.WriteByte(CHR_LF)

   pWriter.Flush()

   os.Exit(0)
}
