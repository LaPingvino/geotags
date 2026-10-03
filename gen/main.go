// Command gen builds places.json from a GeoNames cities dump:
//
//	curl -LO https://download.geonames.org/export/dump/cities15000.zip && unzip cities15000.zip
//	go run ./gen -in cities15000.txt -min 100000 > places.json
//
// Every place gets a hashtag (its ASCII name, lowercased, no spaces), its
// own name as an alias when that differs, a radius estimated from the
// population, and an "ambiguous" flag when the same tag names more than one
// place (or is too short to mean anything on its own).
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type place struct {
	Tag        string   `json:"tag"`
	Aliases    []string `json:"aliases,omitempty"`
	Name       string   `json:"name"`
	Country    string   `json:"country"`
	Population int      `json:"population"`
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	Km         float64  `json:"km"`
	Cell       string   `json:"cell"`
	Ambiguous  bool     `json:"ambiguous,omitempty"`
}

func main() {
	in := flag.String("in", "cities15000.txt", "GeoNames dump (tab-separated)")
	min := flag.Int("min", 100000, "minimum population")
	aliasFile := flag.String("aliases", "aliases.json", "curated extra hashtags per tag")
	noisyFile := flag.String("noisy", "noisy.json", "tags that are also common words or famous elsewhere")
	flag.Parse()
	f, err := os.Open(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	var ps []*place
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<22)
	for sc.Scan() {
		c := strings.Split(sc.Text(), "\t")
		if len(c) < 15 {
			continue
		}
		pop, _ := strconv.Atoi(c[14])
		// Skip sections of a city (Paris' arrondissements, São Paulo's
		// districts): people tag the city, not the section.
		if pop < *min || c[7] == "PPLX" {
			continue
		}
		lat, _ := strconv.ParseFloat(c[4], 64)
		lon, _ := strconv.ParseFloat(c[5], 64)
		tag := compact(c[2])
		if tag == "" {
			continue
		}
		p := &place{Tag: tag, Name: c[1], Country: c[8], Population: pop, Lat: round(lat, 4), Lon: round(lon, 4), Cell: cell(lat, lon)}
		if a := compactKeep(c[1]); a != "" && a != tag {
			p.Aliases = []string{a}
		}
		p.Km = round(math.Min(math.Max(math.Sqrt(float64(pop))/60, 2), 25), 1)
		ps = append(ps, p)
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].Population > ps[j].Population })
	ps = dropSections(ps)
	// Curated aliases go to the most populous place with that tag.
	var extra map[string][]string
	if b, err := os.ReadFile(*aliasFile); err == nil {
		json.Unmarshal(b, &extra)
	}
	done := map[string]bool{}
	for _, p := range ps {
		if done[p.Tag] {
			continue
		}
		done[p.Tag] = true
		for _, a := range extra[p.Tag] {
			if a != p.Tag && !contains(p.Aliases, a) {
				p.Aliases = append(p.Aliases, a)
			}
		}
	}
	count := map[string]int{}
	for _, p := range ps {
		count[p.Tag]++
		for _, a := range p.Aliases {
			count[a]++
		}
	}
	noisy := map[string]bool{}
	if b, err := os.ReadFile(*noisyFile); err == nil {
		var ns []string
		json.Unmarshal(b, &ns)
		for _, n := range ns {
			noisy[n] = true
		}
	}
	for _, p := range ps {
		p.Ambiguous = count[p.Tag] > 1 || len(p.Tag) <= 3 || noisy[p.Tag]
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	fmt.Println("[")
	for i, p := range ps {
		b, _ := json.Marshal(p)
		sep := ","
		if i == len(ps)-1 {
			sep = ""
		}
		fmt.Printf("  %s%s\n", b, sep)
	}
	fmt.Println("]")
	_ = enc
}

// compact lowercases, strips accents and drops everything but letters and
// digits: "São Paulo" → "saopaulo".
func compact(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		r = unicode.ToLower(r)
		if r < 128 && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// compactKeep is compact without stripping accents or non-Latin letters:
// "München" → "münchen", "東京" → "東京".
func compactKeep(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

const alphabet = "23456789cfghjmpqrvwx"

// cell is the #geo cell: the first six Open Location Code characters,
// lowercased (0.05° ≈ 5.5 km).
func cell(lat, lon float64) string {
	lat = math.Min(math.Max(lat, -90), 90-1e-9) + 90
	lon = math.Mod(math.Mod(lon+180, 360)+360, 360)
	var b strings.Builder
	for _, res := range []float64{20, 1, 0.05} {
		la := min(int(math.Floor(lat/res+1e-9)), 19)
		lo := min(int(math.Floor(lon/res+1e-9)), 19)
		b.WriteByte(alphabet[la])
		b.WriteByte(alphabet[lo])
		lat -= float64(la) * res
		lon -= float64(lo) * res
	}
	return b.String()
}

// dropSections removes districts listed as cities of their own ("Paris 17
// Batignolles-Monceau", "Berlin-Mitte"): a place whose tag starts with the
// tag of a bigger place in the same country whose radius covers it.
func dropSections(ps []*place) []*place {
	var out []*place
	for _, p := range ps {
		section := false
		for _, big := range out {
			if big.Country == p.Country && len(big.Tag) >= 4 && len(p.Tag) > len(big.Tag) &&
				strings.HasPrefix(p.Tag, big.Tag) && distKm(big.Lat, big.Lon, p.Lat, p.Lon) <= big.Km {
				section = true
				break
			}
		}
		if !section {
			out = append(out, p)
		}
	}
	return out
}

func distKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	rad := math.Pi / 180
	dLat, dLon := (lat2-lat1)*rad, (lon2-lon1)*rad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*rad)*math.Cos(lat2*rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * R * math.Asin(math.Sqrt(a))
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func round(x float64, d int) float64 { p := math.Pow(10, float64(d)); return math.Round(x*p) / p }
