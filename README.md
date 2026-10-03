# geotags

Hashtags people already use for places and languages, mapped to `#geo` cells and ISO 639-3 codes —
so "what's near me" can be a plain text search on any network.

Made for [Kafumu](https://github.com/LaPingvino/kafumu); free for any project.

## `#geo` cells

A `#geo` cell is the first six characters of a location's [Open Location Code](https://maps.google.com/pluscodes/)
(plus code), lowercased: about 5.5 × 5.5 km. Central Amsterdam is `#geo9f469v`, central Lisbon `#geo8ccgpv`.
A cell is just a word, so posting with `#geo9f469v` makes a post findable by anyone searching that tag, on any
network, without anyone sharing coordinates. Neighbouring cells: offset the centre by ±0.05° and re-encode.

## Files

- **`places.json`** — 6,000+ cities with a population of at least 100,000:
  `tag` (ASCII name, lowercased, no spaces), `aliases` (native spelling and curated common hashtags),
  `name`, `country`, `population`, `lat`, `lon`, `km` (radius estimated from population), `cell`
  (the centre's `#geo` cell), and `ambiguous` when a tag names more than one place or is very short.
  To find the place tags for a cell, take every place whose centre is within `km` of the cell's centre.
- **`towns.json`** — every place of 15,000+ people (`[name, country, lat, lon]`, ~32k), for finding a town by
  name. Not a hashtag list: small-town names are too often ordinary words to use as place tags.
- **`noisy.json`** — tags that are also common words or famous elsewhere (`#paris`, `#nice`, `#reading`),
  always flagged ambiguous.
- **`aliases.json`** — curated extra hashtags (`lisbon` → `lisboa`, `lx`; `mexicocity` → `cdmx`).
- **`languages.json`** — language hashtags (`#esperanto`, `#tokipona`, `#learnjapanese`, `#languageexchange`…)
  → ISO 639-3 `codes` (empty for "any language") and a `weight` for how specific the tag is.

## Regenerating

```bash
curl -LO https://download.geonames.org/export/dump/cities15000.zip && unzip cities15000.zip
go run ./gen -in cities15000.txt -min 100000 > places.json
go run ./towns -in cities15000.txt > towns.json
```

## License

- Data in `places.json` is derived from [GeoNames](https://www.geonames.org/) and licensed under
  [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/). Please credit GeoNames.
- `aliases.json`, `noisy.json`, `languages.json` and the code: CC0 / public domain. Contributions welcome — especially
  hashtags people really use in your city or language community.
