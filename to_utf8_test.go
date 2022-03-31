package ansiart2utf8

import (
	"bufio"
	"io/ioutil"
	"os"
	"testing"
)

//const TEST_DIR = "./test_data"
const TEST_DIR = "./test_data/artwork"

func TestToUTF8(t *testing.T) {

	// BUFFER OUTPUT
	pWriter := bufio.NewWriter(os.Stdout)

	UM := UTF8Marshaller{
		Width:    80,
		MaxBytes: 0,
		Writer:   pWriter,
	}

	// DEBUG LOGGING
	UM.Debug = func(v ...interface{}) (int, error) {

		t.Log(v...)
		return 0, nil
	}

	sFI, oE := ioutil.ReadDir(TEST_DIR)
	if oE != nil {
		t.Fatal(oE.Error())
	}

	for _, FI := range sFI {

		if FI.IsDir() {
			continue
		}

		/*

			TODO:
				- SGR delta encoding
					- determine behavior of unwritten cells
						(i.e. inherit SGR vs B&W)
				- 256 color replacements option
					(to override custom terminal colors)

			PROBLEMS:

			textfiles/artwork
				ufo.ans - missing top

			textfiles/holiday
				wwans53.ans - missing bottom & bag
				vday.ans - missing ll corner
				FIXED: thanks3.ans - missing line

		*/

		if FI.Name() != "thanks3.ans" {
			//continue
		}

		szFile := TEST_DIR + "/" + FI.Name()
		pWriter.WriteString("FILE: " + szFile + "\n")

		if oE := testFile(&UM, szFile); oE != nil {

			t.Fatal(oE.Error())
			return
		}

		pWriter.WriteByte(CHR_LF)
		pWriter.Flush()

		// TODO: reset SGR for errors
	}
}

func testFile(pUM *UTF8Marshaller, szFile string) error {

	pFile, oE := os.Open(szFile)
	if oE != nil {
		return oE
	}
	defer pFile.Close()

	if oE = pUM.Encode(pFile); oE != nil {
		return oE
	}

	return nil
}
