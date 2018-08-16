package main

import (
	"net/http"
	"sync"
)

var (
	varsLock sync.RWMutex
	vars     map[*http.Request]map[string]interface{}
)

func openVars(r *http.Request) {
	varsLock.Lock()
	defer varsLock.Unlock()
	if vars == nil {
		vars = make(map[*http.Request]map[string]interface{})
	}
	vars[r] = map[string]interface{}{}
}

func closeVars(r *http.Request) {
	varsLock.Lock()
	defer varsLock.Unlock()
	delete(vars, r)
}

func getVar(r *http.Request, key string) interface{} {
	varsLock.RLock()
	defer varsLock.RUnlock()
	return vars[r][key]
}

func setVar(r *http.Request, key string, value interface{}) {
	varsLock.Lock()
	defer varsLock.Unlock()
	vars[r][key] = value
}
