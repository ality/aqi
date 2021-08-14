package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestRecent(t *testing.T) {
	test(t, &recent{}, "testdata/060670010.json")
}

func TestForecast(t *testing.T) {
	test(t, &forecast{}, "testdata/airnowapi.json")
}

func test(t *testing.T, l loader, file string) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	want, err := ioutil.ReadFile(filepath.Join(dir, file+".golden"))
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(filepath.Join(dir, file))
	if err != nil {
		t.Fatal(err)
	}
	w, err := l.Load(f)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_, err = w.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	have := buf.Bytes()
	if bytes.Compare(have, want) != 0 {
		t.Logf("have (%d):\n%s", len(have), have)
		t.Logf("want (%d):\n%s", len(want), want)
		t.FailNow()
	}
}

type loader interface {
	Load(rd io.Reader) (io.WriterTo, error)
}
