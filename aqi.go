package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: aqi [lat long]\n")
	os.Exit(1)
}

func main() {
	log.SetPrefix("aqi: ")
	log.SetFlags(0)

	var lat, long float64
	switch arg := os.Args[1:]; len(arg) {
	case 0:
		// TODO: read local coordinates from /lib/sky/here; see astro(1)
		lat, long = 38.57668336927141, -121.49356942613394
	case 2:
		f64 := func(s string) float64 {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				usage()
			}
			return f
		}
		lat, long = f64(arg[0]), f64(arg[1])
	default:
		// TODO: find coordinates using the arcgis world geocoding service
		// https://developers.arcgis.com/rest/location-based-services/
		usage()
	}

	var (
		fs = []fetcher{
			&recent{},
			&forecast{},
		}
		cs []chan result
		wg sync.WaitGroup
	)
	for _, f := range fs {
		f := f
		c := make(chan result, 1)
		wg.Add(1)
		cs = append(cs, c)
		go func() {
			t, err := f.Fetch(lat, long)
			c <- result{t, err}
			wg.Done()
		}()
	}
	wg.Wait()

	w := bufio.NewWriter(os.Stdout)
	for i, c := range cs {
		r := <-c
		if r.err != nil {
			log.Print(r.err)
			continue
		}
		if i > 0 {
			w.WriteString("\n")
		}
		r.t.WriteTo(w)
	}
	w.Flush()
}

type result struct {
	t   io.WriterTo
	err error
}

type fetcher interface {
	Fetch(lat, long float64) (io.WriterTo, error)
}
