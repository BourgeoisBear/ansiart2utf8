package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	ansi "github.com/BourgeoisBear/ansiart2utf8"
)

func main() {

	var oErr error

	// ERROR LOGGING
	pLogErr := log.New(os.Stderr, "", log.Lshortfile)
	defer func() {

		if oErr != nil {

			pLogErr.Output(2, oErr.Error())
			fmt.Fprint(os.Stdout, "\x1b[0m")
			os.Exit(1)
		}
	}()

	runtime.GOMAXPROCS(1)

	const SZ_HELP_PREFIX = `
ansiart2utf8
	Converts ANSI art to UTF-8 encoding, expands cursor forward ESC sequences
	into spaces, wraps/resets at a specified line width, sends result to STDOUT.

	Leave the [FILE] parameter empty to read from STDIN.

USAGE: ansiart2utf8 [OPTION]... [FILE]...

OPTIONS
`

	// HELP MESSAGE
	flag.Usage = func() {

		fmt.Fprint(os.Stdout, SZ_HELP_PREFIX)
		flag.PrintDefaults()
		fmt.Fprint(os.Stdout, "\n")
	}

	UM := ansi.UTF8Marshaller{}

	// COMMAND PARAMETERS
	pbDebug := flag.Bool("debug", false, `DEBUG MODE: line numbering + pipe @ \n`)
	flag.BoolVar(&UM.Translate2Xterm256, "x", false, "ANSI TO XTERM-256 COLOR SUBSTITUTION\n  (to overcome strange terminal color scheme palettes)")

	flag.UintVar(&UM.Width, "w", 80, "LINE WRAP WIDTH")
	flag.UintVar(&UM.MaxBytes, "bytes", 0, "MAXIMUM OUTPUT BYTES PER-ROW (0 = NO LIMIT)")

	flag.Parse()

	if !flag.Parsed() {

		oErr = errors.New("Invalid Parameters")
		return
	}

	if UM.Width < 1 {

		oErr = errors.New("LINE WIDTH must be > 0")
		return
	}

	// DEBUG LOGGING
	fnDebug := func(v ...interface{}) (int, error) {

		if *pbDebug {
			return fmt.Println(v...)
		}

		return 0, nil
	}

	// BUFFER OUTPUT
	pWriter := bufio.NewWriter(os.Stdout)
	UM.Writer = pWriter

	if *pbDebug {
		UM.Debug = fnDebug
	}

	// PROCESS INPUT FILES
	arFiles := flag.Args()
	if len(arFiles) == 0 {
		arFiles = append(arFiles, "-")
	}

	for _, szFname := range arFiles {

		if szFname == "-" {

			fnDebug("PROCESSING STDIN")
			oErr = UM.Encode(os.Stdin)
			if oErr != nil {
				return
			}

		} else {

			pF, oE2 := os.Open(szFname)
			if oE2 != nil {
				oErr = fmt.Errorf("UNABLE TO OPEN %s: %s", szFname, oE2.Error())
				return
			}

			fnDebug("PROCESSING ", szFname)
			oErr = UM.Encode(pF)
			pF.Close()
			if oErr != nil {
				return
			}
		}

		pWriter.WriteByte(ansi.CHR_LF)
		pWriter.Flush()
	}

	os.Exit(0)
}
