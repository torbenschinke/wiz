package pack

import (
	io2 "github.com/torbenschinke/wiz/io"
	"io"
	"os"
	"sync"
)

type SingleFileReader struct {
	file   *os.File
	offset int64
	mutex  sync.Mutex
}

func NewSingleFileReader(file io2.File) (*SingleFileReader, error) {
	f, err := os.Open(string(file))
	if err != nil {
		return nil, err
	}
	return &SingleFileReader{file: f}, nil
}

func (r *SingleFileReader) Read(id uint32, offset int64, buf []byte) (int, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.offset != offset {
		newOffset, err := r.file.Seek(offset, io.SeekStart)
		r.offset = newOffset
		if err != nil {
			return 0, err
		}
	}
	return r.file.Read(buf)
}
