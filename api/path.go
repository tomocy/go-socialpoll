package main

import "strings"

type apiPath struct {
	pathBeforeID string
	id           string
}

const pathSeparator = "/"

func newAPIPath(path string) *apiPath {
	pathBeforeID, id := separateIntoPathBeforeIDAndID(path)
	return &apiPath{
		pathBeforeID: pathBeforeID,
		id:           id,
	}
}

func separateIntoPathBeforeIDAndID(p string) (string, string) {
	trimmedPath := strings.Trim(p, pathSeparator)
	pathPieces := strings.Split(trimmedPath, pathSeparator)
	var id string
	if 2 <= len(pathPieces) {
		id = pathPieces[len(pathPieces)-1]
		trimmedPath = strings.Join(pathPieces[:len(pathPieces)-1], pathSeparator)
	}
	return trimmedPath, id
}

func (p apiPath) hasID() bool {
	return 1 <= len(p.id)
}
