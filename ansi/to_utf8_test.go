package ansi

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

const TESTS_DIR = "../test_data"

func getTestFiles() ([]string, error) {

	pFile, E := os.Open(TESTS_DIR)
	if E != nil {
		return nil, E
	}
	defer pFile.Close()

	sFI, E := pFile.ReadDir(0)
	if E != nil {
		return nil, E
	}

	ret := make([]string, 0, len(sFI))
	for _, FI := range sFI {
		if !FI.IsDir() && strings.HasSuffix(FI.Name(), ".ans") {
			ret = append(ret, TESTS_DIR+"/"+FI.Name())
		}
	}

	return ret, nil
}

func TestToUTF8(t *testing.T) {

	sFiles, E := getTestFiles()
	if E != nil {
		t.Fatal(E)
	}

	pLogDebug := log.New(os.Stdout, "", 0)

	for _, fname := range sFiles {

		if pFile, E := os.Open(fname); E != nil {

			E = fmt.Errorf("FILE: %s, ERROR: %s", fname, E.Error())
			t.Error(E)

		} else {

			t.Logf("TESTING: %s\n", fname)
			if E = ToUTF8(pFile, os.Stdout, 80, 0, pLogDebug); E != nil {
				t.Error(E)
			}

			pFile.Close()
		}
	}
}
