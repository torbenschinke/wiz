package io

import (
	"encoding/binary"
	"io"
)

type DataIn interface {
	ReadUInt32LE() uint32
	ReadByte() byte
	Read(buf []byte) (int, error)
	Error() error
}

type DataInReader struct {
	delegate    io.Reader
	tmp1        []byte
	tmp2        []byte
	tmp4        []byte
	tmp8        []byte
	lastErr     error
	readLastErr int
}

func NewDataIn(r io.Reader) *DataInReader {
	return &DataInReader{delegate: r, tmp1: make([]byte, 1), tmp2: make([]byte, 2), tmp4: make([]byte, 4), tmp8: make([]byte, 8)}
}

func (r *DataInReader) ReadUInt32LE() uint32 {
	if r.lastErr != nil {
		return 0
	}
	n, err := r.delegate.Read(r.tmp4)
	if err != nil {
		r.lastErr = err
		r.readLastErr = n
		return 0
	}
	return binary.LittleEndian.Uint32(r.tmp4)
}

func (r *DataInReader) ReadByte() byte {
	if r.lastErr != nil {
		return 0
	}
	n, err := r.delegate.Read(r.tmp1)
	if err != nil {
		r.lastErr = err
		r.readLastErr = n
		return 0
	}
	return r.tmp1[0]
}

func (r *DataInReader) Read(buf []byte) (int, error) {
	if r.lastErr != nil {
		return r.readLastErr, r.lastErr
	}
	n, err := r.delegate.Read(buf)
	if err != nil {
		r.lastErr = err
		r.readLastErr = n
		return n, err
	}
	return n, nil
}

func (r *DataInReader) Error() error {
	return r.lastErr
}
