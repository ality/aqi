package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/doc"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// http://files.airnowtech.org/?prefix=airnow/today/
// https://docs.airnowapi.org/

type forecast struct{}

func (f *forecast) Fetch(lat, long float64) (io.WriterTo, error) {
	// TODO: determine state code from coordinates
	v := url.Values{}
	v.Set("latitude", strconv.FormatFloat(lat, 'f', -1, 64))
	v.Set("longitude", strconv.FormatFloat(long, 'f', -1, 64))
	v.Set("maxDistance", "50")
	v.Set("stateCode", "CA")
	res, err := http.PostForm("https://airnowgovapi.com/reportingarea/get", v)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch forecast: %v", err)
	}
	defer res.Body.Close()
	return f.Load(res.Body)
}

func (f *forecast) Load(rd io.Reader) (io.WriterTo, error) {
	b, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, fmt.Errorf("cannot read forecast: %v", err)
	}
	var r records
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("cannot unmarshal forecast: %v", err)
	}
	return &r, nil
}

type records []record

func (rs *records) WriteTo(wr io.Writer) (int64, error) {
	// jq -Mr 'map(select(.dataType == "F")) | .[0].discussion' | fmt

	var s string
	for _, r := range *rs {
		if r.DataType == "F" {
			s = r.Discussion
			break
		}
	}
	var b bytes.Buffer
	doc.ToText(&b, s, "", "", 70)
	return b.WriteTo(wr)
}

type record struct {
	DataType   string
	Discussion string
}
