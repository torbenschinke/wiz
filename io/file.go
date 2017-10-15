package io

import "os"

type File string

func (f File) IsFile() bool {
	if stat, err := os.Stat(string(f)); err != nil {
		return false
	} else {
		return stat.Mode().IsRegular()
	}
}

func (f File) IsDir() bool {
	if stat, err := os.Stat(string(f)); err != nil {
		return false
	} else {
		return stat.Mode().IsDir()
	}
}

func (f File) Exists() bool {
	if _, err := os.Stat(string(f)); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
