package hcplog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogList(t *testing.T) {
	r := bytes.NewReader(helperLoad(t, t.Name()+".json"))
	logFiles, err := ParseLogList(r)
	assert.Equal(t, 123, 123, "they should be equal")
	assert.Nil(t, err, "Cant parse log files payload")
	fmt.Printf("Log files: %v", logFiles)
}

// See https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859
func helperLoad(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
