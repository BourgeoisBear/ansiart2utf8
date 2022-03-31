package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	ansi "github.com/BourgeoisBear/ansiart2utf8"
)

func main() {

	var (
		oErr  error
		pFile *os.File = nil
	)

	// ERROR LOGGING
	pLogErr := log.New(os.Stderr, "", log.Lshortfile)
	defer func() {

		if oErr != nil {

			pLogErr.Output(2, oErr.Error())
			os.Exit(1)
		}
	}()

	runtime.GOMAXPROCS(1)

	const SZ_HELP_PREFIX = `
ansiart2utf8
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
	pszInput := flag.String("f", "-", "INPUT FILENAME, OR \"-\" FOR STDIN")
	pbDebug := flag.Bool("d", false, "DEBUG MODE: LINE NUMBERING + PIPE @ \\n")
	puiWidth := flag.Uint("w", 80, "LINE WIDTH")
	pnRowBytes := flag.Uint("bytes", 0, "MAXIMUM OUTPUT BYTES PER-ROW (0 = NO LIMIT)")

	flag.Parse()

	if !flag.Parsed() {

		oErr = errors.New("Invalid Parameters")
		return
	}

	if *puiWidth < 1 {

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

	// GET FILE HANDLE
	if strings.Compare(*pszInput, "-") == 0 {

		pFile = os.Stdin

	} else {

		if pFile, oErr = os.Open(*pszInput); oErr != nil {
			return
		}
		defer pFile.Close()

		fnDebug("FILE: ", *pszInput)
	}

	// BUFFER OUTPUT
	pWriter := bufio.NewWriter(os.Stdout)

	UM := ansi.UTF8Marshaller{
		Width:    *puiWidth,
		MaxBytes: *pnRowBytes,
		Writer:   pWriter,
	}

	if *pbDebug {
		UM.Debug = fnDebug
	}

	if oErr = UM.Encode(pFile); oErr != nil {
		return
	}

	pWriter.WriteByte(ansi.CHR_LF)
	pWriter.Flush()

	os.Exit(0)
}
