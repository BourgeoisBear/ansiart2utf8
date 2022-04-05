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
		Width:              80,
		MaxBytes:           0,
		Writer:             pWriter,
		Translate2Xterm256: true,
		FakeEsc:            false,
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

			TODO: PROBLEMS

			textfiles/holiday
				wwans53.ans - missing bottom & bag
				vday.ans - missing ll corner
				FIXED: thanks3.ans - missing line

		*/

		if FI.Name() != "fruit.ans" {
			// continue
		}

		szFile := TEST_DIR + "/" + FI.Name()
		pWriter.WriteString("FILE: " + szFile + "\n")

		if oE := testFile(&UM, szFile); oE != nil {

			t.Fatal(oE.Error())
			return
		}

		pWriter.WriteByte(CHR_LF)
		pWriter.Flush()
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
