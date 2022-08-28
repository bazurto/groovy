package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func cacheAny(cacheFileName string, v any, f func() (any, error)) error {

	// from cache
	if fileExists(cacheFileName) {
		file, err := os.Open(cacheFileName)
		if err == nil {
			defer file.Close()
			err = json.NewDecoder(file).Decode(v)
			if err == nil {
				return nil
			}
		}
	}

	// no cache
	tmp, err := f()
	if err != nil {
		return fmt.Errorf("cache error, function returned error: %s", err)
	}
	b, err := json.Marshal(tmp)
	if err != nil {
		return fmt.Errorf("cache error marshaling: %s", err)
	}

	err = json.Unmarshal(b, v) // assign to parameter
	if err != nil {
		return fmt.Errorf("cache error unmarshaling error: %s", err)
	}

	// write to cache
	ioutil.WriteFile(cacheFileName, b, 0660)

	return nil
}

func cache(cachefile string, f func() error) error {
	if fileExists(cachefile) {
		return nil
	}
	return f()
}
