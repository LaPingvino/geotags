// Command towns builds towns.json: every place of 15,000+ people from
// GeoNames as [name, country, lat, lon] — for finding a town by name. It
// is deliberately not a hashtag list: small-town names are too often
// ordinary words to use as place tags (see places.json for those).
//
//	go run ./towns -in cities15000.txt > towns.json
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

func main() {
	in := flag.String("in", "cities15000.txt", "GeoNames dump")
	flag.Parse()
	f, err := os.Open(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<22)
	var rows [][]any
	for sc.Scan() {
		c := strings.Split(sc.Text(), "\t")
		if len(c) < 15 || c[7] == "PPLX" {
			continue
		}
		lat, _ := strconv.ParseFloat(c[4], 64)
		lon, _ := strconv.ParseFloat(c[5], 64)
		rows = append(rows, []any{c[1], c[8], math.Round(lat*1e3) / 1e3, math.Round(lon*1e3) / 1e3})
	}
	b, _ := json.Marshal(rows)
	os.Stdout.Write(b)
}
