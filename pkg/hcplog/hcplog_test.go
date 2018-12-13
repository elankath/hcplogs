package hcplog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogList(t *testing.T) {
	r := bytes.NewReader(helperLoad(t, t.Name()+".json"))
	logFiles, err := ParseLogList(r)
	log.Printf("%s: Failed: %v", t.Name(), err)
	if assert.NoError(t, err) {
		fmt.Printf("%s: Log files: %v\n", t.Name(), logFiles)
		assert.NotNil(t, logFiles, "logFiles must not be nil")
	} else {
		log.Printf("%s: Failed: %v", t.Name(), err)
	}
}

// See https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859
func helperLoad(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name) // relative path
	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
