package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type recent struct{}

func (r *recent) Fetch(lat, lon float64) (io.WriterTo, error) {
	s := stationID(lat, lon)
	u := fmt.Sprintf("https://an_gov_data.s3.amazonaws.com/Sites/%s.json", s)
	res, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch recent: %v", err)
	}
	defer res.Body.Close()
	return r.Load(res.Body)
}

func (r *recent) Load(rd io.Reader) (io.WriterTo, error) {
	b, err := ioutil.ReadAll(rd)
	if err != nil {
		return nil, fmt.Errorf("cannot read recent: %v", err)
	}
	var s station
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("cannot unmarshal recent: %v", err)
	}
	return &s, nil
}

type station struct {
	SiteName     string
	StationID    string
	Monitors     []monitor
	UTCDateTimes []ztime
}

func (s *station) WriteTo(wr io.Writer) (n int64, err error) {
	print := func(a ...interface{}) {
		if err == nil {
			var m int
			m, err = fmt.Fprintf(wr, "%-25s\t%-17s\t%-17s\t%s\n", a...)
			n += int64(m)
		}
	}
	print("measurement start time", "pm25 aqi", "pm10 aqi", "ozone aqi")
	const limit = 12
	o := len(s.UTCDateTimes) - limit
	if o < 0 {
		o = 0
	}
	// NOTE: Monitors and UTCDateTimes arrive pre-sorted
	for i, t := range s.UTCDateTimes[o:] {
		p := make(map[name]value)
		for _, m := range s.Monitors {
			p[m.Name] = value{m.AQI[o+i], m.Conc[o+i], m.Unit}
		}
		print(t, p["pm25"], p["pm10"], p["ozone"])
	}
	return
}

type monitor struct {
	Name name `json:"parameterName"`
	Unit unit `json:"concUnit"`
	AQI  []float64
	Conc []float64
}

type name string

func (n *name) UnmarshalJSON(b []byte) error {
	s := string(b)
	switch s {
	case `"O3"`:
		*n = "ozone"
	case `"PM2.5 - Principal"`:
		*n = "pm25"
	case `"PM10 - Principal"`:
		*n = "pm10"
	default:
		*n = name(s)
	}
	return nil
}

type unit string

func (u *unit) UnmarshalJSON(b []byte) error {
	s := string(b)
	switch s {
	case `"PPB"`:
		*u = "ppb"
	case `"UG/M3"`:
		*u = "µg/m³"
	default:
		*u = unit(s)
	}
	return nil
}

type ztime time.Time // zulu time (UTC)

func (z *ztime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	t, err := time.Parse(`"2006-01-02 15:04:05"`, string(b))
	*z = ztime(t)
	return err
}

func (z ztime) String() string {
	return time.Time(z).Local().Format(time.RFC3339)
}

type value struct {
	aqi  float64
	conc float64
	unit unit
}

func (v value) String() string {
	a, c := "-", "-"
	if v.aqi > 0 {
		a = strconv.Itoa(int(v.aqi))
	}
	if v.conc > 0 {
		c = strconv.Itoa(int(v.conc))
	}
	return fmt.Sprintf("%c %s (%s %s)", v.sym(), a, c, v.unit)
}

func (v value) sym() rune {
	switch a := v.aqi; {
	case a > 300:
		return 'H' // hazardous
	case a > 200:
		return 'V' // very unhealthy
	case a > 150:
		return 'U' // unhealthy
	case a > 100:
		return 'S' // unhealthy for sensitve groups
	case a > 50:
		return 'M' // moderate
	default:
		return 'G' // good
	}
}

func stationID(lat, lon float64) string {
	// TODO: find nearest station by latitude and longitude
	//
	// fetch a list of stations from the airnow.gov api.
	//
	// calculate the distance between our location and each
	// station (by solving the inverse geodesic problem on the
	// wgs84 ellipsoid) and sort to find the nearest station.
	//
	// the above algorithm is O(n) and fine for a small set of
	// stations. if the set is large, however, we should create
	// a vantage-point tree in O(n log n) time to get O(log n)
	// queries.

	// the permanent air quality monitor for downtown sacramento is
	// located on the roof of the air resources board monitoring and
	// laboratory division building at 1309 t street.
	// https://ww3.arb.ca.gov/qaweb/iframe_site.php?s_arb_code=34305
	return "060670010"
}
