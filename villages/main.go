// Command villages builds villages.tsv.gz: every place of 1,000+ people
// from GeoNames, neighbourhoods (PPLX) included, as
// "name\tCC\tlat\tlon\tkind\tpopulation" lines (kind t = town, n =
// neighbourhood), for naming the area you're in ("Lunteren", "Areeiro,
// Lisbon"), never as hashtags.
//
//	go run ./villages -in cities1000.txt | gzip -9 > villages.tsv.gz
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	in := flag.String("in", "cities1000.txt", "GeoNames dump")
	flag.Parse()
	f, err := os.Open(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<22)
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for sc.Scan() {
		c := strings.Split(sc.Text(), "\t")
		if len(c) < 15 {
			continue
		}
		lat, _ := strconv.ParseFloat(c[4], 64)
		lon, _ := strconv.ParseFloat(c[5], 64)
		pop, _ := strconv.Atoi(c[14])
		kind := "t"
		if c[7] == "PPLX" {
			kind = "n"
		}
		fmt.Fprintf(w, "%s\t%s\t%.3f\t%.3f\t%s\t%d\n", c[1], c[8], lat, lon, kind, pop)
	}
}
