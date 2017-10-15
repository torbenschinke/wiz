package wiz

import (
	"github.com/torbenschinke/wiz/io"
	"io/ioutil"
	"testing"
)

func TestOpen(t *testing.T) {

	tmpFile, _ := ioutil.TempFile("", "")
	defer tmpFile.Close()
	_, err := Open(io.File(tmpFile.Name()))
	if err != nil {
		t.Fatal(err)
	}
}
