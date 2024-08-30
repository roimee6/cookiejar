package cookiejar

import (
	"net/http"
)

type Storage interface {
	Set(string, map[string]Entry)
	Get(string) map[string]Entry
	Delete(string)
}

func NewFileJar(filename string, o *Options) (http.CookieJar, error) {
	store := &FileDrive{
		filename: filename,
		entries:  make(map[string]map[string]Entry),
	}
	store.readEntries()

	return New(store, o)
}

func NewEntriesJar(o *Options) (EntriesDrive, http.CookieJar, error) {
	store := &EntriesDrive{
		entries: make(map[string]map[string]Entry),
	}

	jar, err := New(store, o)
	return *store, jar, err
}
