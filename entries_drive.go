package cookiejar

type EntriesDrive struct {
	entries map[string]map[string]Entry
}

func (e *EntriesDrive) Set(key string, val map[string]Entry) {
	e.entries[key] = val
}

func (e *EntriesDrive) Get(key string) map[string]Entry {
	return e.entries[key]
}

func (e *EntriesDrive) Delete(key string) {
	delete(e.entries, key)
}

func (e *EntriesDrive) GetEntries() map[string]map[string]Entry {
	return e.entries
}