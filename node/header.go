package node

import "github.com/torbenschinke/wiz/io"

/**
Format (fixed 14 byte):
				  name                value        length         type
			---------------------|---------------|-----------|----------------
				 node type             0x00         1 byte      byte
				 magic             [77 69 7a 63]    4 byte      [4]byte
				 sub format magic    [* * * *]      4 byte      [4]byte
				 version               0x03         4 byte      uint32
				 encryption             *           1 byte      byte
*/
type Header struct {
	Type       NodeType
	Magic      [4]byte
	SubMagic   [4]byte
	Version    uint32
	Encryption Encryption
}

func (h Header) GetType() NodeType {
	return h.Type
}

func (h Header) Read(in io.DataIn) error {
	h.Type = NodeType(in.ReadByte())
	in.Read(h.Magic[:])
	in.Read(h.SubMagic[:])
	h.Version = in.ReadUInt32BE()
	h.Encryption = Encryption(in.ReadByte())
	return in.Error()
}

type Encryption byte

const (
	ENC_NONE        = 0x00
	ENC_AES_256_CTR = 0x01
)
