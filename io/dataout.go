package io

import "io"

type DataOut interface {
	WriteUInt32LE(val uint32)
	WriteUInt32BE(val uint32)
	WriteByte(val byte)
	Write(buf []byte) (int, error)
	Error() error
}

type DataOutWriter struct {
	delegate    io.Writer
	tmp1        []byte
	tmp2        []byte
	tmp4        []byte
	tmp8        []byte
	lastErr     error
	readLastErr int
}

func (w *DataOutWriter) WriteUInt32LE(val uint32) {

}

func (w *DataOutWriter) WriteUInt32BE(val uint32) {

}

func (w *DataOutWriter) WriteByte(val byte) {
	if w.lastErr != nil {
		return
	}
	w.tmp1[0] = val
	_, w.lastErr = w.delegate.Write(w.tmp1)
}

func (w *DataOutWriter) Write(buf []byte) (int, error) {
	if w.lastErr != nil{
		return w.
	}
}

func (w *DataOutWriter) Error() error {
	return w.lastErr
}
