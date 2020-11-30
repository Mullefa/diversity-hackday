package internal

import (
	"net/url"
	"testing"
)

func parseUrlUnsafe(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return u
}

func TestProfilesFromBylineHtml(t *testing.T) {

	webUrl := parseUrlUnsafe("https://theguardian.com")

	tests := []struct {
		name       string
		webUrl     *url.URL
		bylineHtml string
		expected   []*url.URL
	}{
		{

			"single byline",
			webUrl,
			"<a href=\"profile/barneyronay\">Barney Ronay</a> at the Emirates Stadium",
			[]*url.URL{parseUrlUnsafe("https://theguardian.com/profile/barneyronay")},
		},
		{
			"multiple bylines",
			webUrl,
			"<a href=\"profile/helen-davidson\">Helen Davidson</a> (now) and <a href=\"profile/luke-henriques-gomes\">Luke Henriques-Gomes</a> and <a href=\"profile/amy-remeikis\">Amy Remeikis</a> (earlier)",
			[]*url.URL{
				parseUrlUnsafe("https://theguardian.com/profile/helen-davidson"),
				parseUrlUnsafe("https://theguardian.com/profile/luke-henriques-gomes"),
				parseUrlUnsafe("https://theguardian.com/profile/amy-remeikis"),
			},
		},
		{
			"no bylines",
			webUrl,
			"Associated Press in New York",
			[]*url.URL{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := JournalistInfoFromByline(test.bylineHtml, test.webUrl)
			if len(actual) != len(test.expected) {
				t.Errorf("wanted: %v, got: %v", test.expected, actual)
				return
			}
			for i := range actual {
				if actual[i].ProfileURL.String() != test.expected[i].String() {
					t.Errorf("(index %d) wanted: %v, got: %v", i, test.expected, actual)
				}
			}
		})
	}

}
