package pack

type PackCollectionReader interface {
	//reads from the given pack file at the given offset into the given buffer
	Read(id uint32, offset int64, buf []byte) (int, error)
}

type PackReader struct {
	packCollection PackCollectionReader
	id             uint32
	offset         int64
}

func NewPackReader(packCollection PackCollectionReader, id uint32) *PackReader {
	return &PackReader{packCollection, id, 0}
}

func (r *PackReader) Read(buf []byte) (int, error) {
	n, err := r.packCollection.Read(r.id, r.offset, buf)
	r.offset += int64(n)
	return n, err
}
