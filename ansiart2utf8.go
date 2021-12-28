package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BourgeoisBear/ansiart2utf8/ansi"
)

func main() {

	// ERROR LOGGING
	var oErr error
	pLogErr := log.New(os.Stderr, "", log.Lshortfile)
	defer func() {
		if oErr != nil {
			pLogErr.Output(2, oErr.Error())
			os.Exit(1)
		}
	}()

	const SZ_HELP_PREFIX = `
ansiart2utf8 VERSION 0.4
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

	if flag.Parse(); !flag.Parsed() {
		oErr = errors.New("Invalid Parameters")
		return
	}

	// DEBUG LOGGING
	var pLogDebug *log.Logger
	if *pbDebug {
		pLogDebug = log.New(os.Stdout, "", 0)
	}

	// GET FILE HANDLE
	var pFile *os.File
	if strings.Compare(*pszInput, "-") == 0 {

		pFile = os.Stdin

	} else {

		if pFile, oErr = os.Open(*pszInput); oErr != nil {
			oErr = fmt.Errorf("FILE: %s, ERROR: %s", *pszInput, oErr.Error())
			return
		}
		defer pFile.Close()
	}

	if oErr = ansi.ToUTF8(pFile, os.Stdout, *puiWidth, *pnRowBytes, pLogDebug); oErr != nil {
		return
	}

	os.Exit(0)
}
