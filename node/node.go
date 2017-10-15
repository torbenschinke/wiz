package node

import "github.com/torbenschinke/wiz/io"

type Node interface {
	/*
		Returns the actual node type
	*/
	GetType() NodeType
	/*
		Reads the node data by interpreting the given in stream
	*/
	Read(in io.DataIn) error
}

type NodeType byte

const (
	N_Header NodeType = 0x00
)
