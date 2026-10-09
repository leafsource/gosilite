package useragent

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		ua         string
		browser, o string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0 Safari/537.36 Edg/120.0", "Edge", "Windows"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Firefox/121.0", "Firefox", "macOS"},
		{"curl/8.7.1", "curl", ""},
		{"", "", ""},
	}
	for _, c := range cases {
		b, o := Parse(c.ua)
		if b != c.browser || o != c.o {
			t.Fatalf("ua=%q: got (%q,%q) want (%q,%q)", c.ua, b, o, c.browser, c.o)
		}
	}
}
