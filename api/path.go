package main

import "strings"

type path struct {
	path string
	id   string
}

const pathSeparator = "/"

func newPath(p string) *path {
	trimmedPath := strings.Trim(p, pathSeparator)
	pathPieces := strings.Split(trimmedPath, pathSeparator)
	var id string
	if 2 <= len(pathPieces) {
		id = pathPieces[len(pathPieces)-1]
		trimmedPath = strings.Join(pathPieces[:len(pathPieces)-1], pathSeparator)
	}

	return &path{
		path: trimmedPath,
		id:   id,
	}
}

func (p path) hasID() bool {
	return 1 <= len(p.id)
}
