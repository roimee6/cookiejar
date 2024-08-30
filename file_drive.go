package cookiejar

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type FileDrive struct {
	filename string
	entries  map[string]map[string]Entry
}

var (
	errMissingRecord = errors.New("missing cookie record")
)

func (f *FileDrive) Set(key string, val map[string]Entry) {
	f.entries[key] = val
	f.saveEntries()
}

func (f *FileDrive) Get(key string) map[string]Entry {
	return f.entries[key]
}

func (f *FileDrive) Delete(key string) {
	delete(f.entries, key)
	f.saveEntries()
}

func (f *FileDrive) saveEntries() {

	var fp *os.File
	if checkFileIsExist(f.filename) {
		fp, _ = os.OpenFile(f.filename, os.O_WRONLY|os.O_TRUNC, 0600)
	} else {
		fp, _ = os.Create(f.filename)
	}

	defer func(fp *os.File) {
		_ = fp.Close()
	}(fp)

	var records [][]string

	for _, sub := range f.entries {
		for _, e := range sub {
			records = append(records, e.Record())
		}
	}

	writer := csv.NewWriter(fp)
	writer.Comma = '\t'
	_ = writer.WriteAll(records)
}

func (f *FileDrive) readEntries() {
	cookies := f.readAll()

	for _, cookie := range cookies {
		key := jarKey(cookie.Domain, nil)

		e := Entry{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			HttpOnly: cookie.HttpOnly,
			Secure:   cookie.Secure,
			Expires:  cookie.Expires,
		}

		if _, ok := f.entries[key]; !ok {
			f.entries[key] = make(map[string]Entry)
		}
		f.entries[key][e.id()] = e
	}
}

func (f *FileDrive) readAll() []*http.Cookie {

	var fp *os.File
	if checkFileIsExist(f.filename) {
		fp, _ = os.OpenFile(f.filename, os.O_RDONLY, 0)
	} else {
		fp, _ = os.Create(f.filename)
	}
	defer fp.Close()

	reader := csv.NewReader(fp)
	reader.Comma = '\t'
	//reader.Comment = '#'
	reader.FieldsPerRecord = -1
	//reader.TrimLeadingSpace = true

	cookies := make([]*http.Cookie, 0)

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			continue
		}

		cookie, err := f.convert(record)
		if err != nil {
			continue
		}

		cookies = append(cookies, cookie)
	}

	return cookies
}

func (f *FileDrive) convert(record []string) (cookie *http.Cookie, err error) {
	if len(record) != 7 {
		return nil, errMissingRecord
	}

	var (
		httpOnlyPrefix       = "#HttpOnly_"
		httpOnlyPrefixLength = strings.Count(httpOnlyPrefix, "") - 1
		recordByte           = []byte(record[0])
		domain               = record[0]
		httpOnly             = false
	)

	if string(recordByte[0:httpOnlyPrefixLength]) == httpOnlyPrefix {
		domain = string(recordByte[httpOnlyPrefixLength:])
		httpOnly = true
	} else {
		if recordByte[0] == '#' {
			return nil, errMissingRecord
		}
	}

	_, path, _secure, _expires, name, value := record[1], record[2], record[3], record[4], record[5], record[6]

	cookie = &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   domain,
		HttpOnly: httpOnly,
	}

	secure, err := strconv.ParseBool(_secure)
	if err == nil {
		cookie.Secure = secure
	}

	expires, err := strconv.ParseInt(_expires, 10, 64)
	if err != nil {
		return nil, err
	}

	exptime := time.Unix(expires, 0)

	// zero probably means that it never expires,
	// or that it is good for as long as this session lasts.
	if expires > 0 {
		cookie.Expires = exptime
	}

	return
}

func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
