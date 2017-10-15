package wiz

import (
	"github.com/torbenschinke/wiz/io"
	"github.com/torbenschinke/wiz/node"
	"github.com/torbenschinke/wiz/pack"
)

type Wiz struct {
	//either a single file or a directory
	db pack.PackCollectionReader
}

func Open(file io.File) (*Wiz, error) {
	wiz := &Wiz{}
	if file.IsFile() {
		db, err := pack.NewSingleFileReader(file)
		if err != nil {
			return nil, err
		}
		wiz.db = db
	} else {
		panic("folder not yet implemented")
	}

	pack := pack.NewPackReader(wiz.db, 0)
	dataIn := io.NewDataIn(pack)
	header := node.Header{}
	err := header.Read(dataIn)
	if err != nil {
		return nil, err
	}
	if header.GetType() != node.N_Header {
		return nil, node.ErrNotAWizContainer
	}
	return wiz, nil
}
