package internal

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Models the subset of a CAPI search response required by this application.
type CAPISearchResponse struct {
	Response struct {
		Total       int `json:"total"`
		StartIndex  int `json:"startIndex"`
		PageSize    int `json:"pageSize"`
		CurrentPage int `json:"currentPage"`
		Pages       int `json:"pages"`
		Results     []struct {
			ID                 string    `json:"id"`
			WebURL             string    `json:"webUrl"` // TODO: can this be URL.ProfileURL ?
			WebPublicationDate time.Time `json:"webPublicationDate"`
			Fields             struct {
				BylineHTML string `json:"bylineHtml"`
			}
		} `json:"results"`
	} `json:"response"`
}

type CAPISearchPaginator struct {
	client *http.Client
	URL    *url.URL
	init   bool
	res    CAPISearchResponse
}

func NewCAPISearchPaginator(client *http.Client, url *url.URL) *CAPISearchPaginator {
	return &CAPISearchPaginator{client: client, URL: url}
}

func (paginator *CAPISearchPaginator) HasNext() bool {
	return !paginator.init || paginator.res.Response.CurrentPage < paginator.res.Response.Pages
}

func (paginator *CAPISearchPaginator) Next() (CAPISearchResponse, error) {
	paginator.init = true
	req := &http.Request{URL: paginator.URL}
	err := doJsonRequest(paginator.client, req, &paginator.res)
	if err == nil {
		query := paginator.URL.Query()
		page, err1 := strconv.Atoi(query.Get("page"))
		if err1 != nil {
			page = 1
		}
		query.Set("page", strconv.Itoa(page+1))
		paginator.URL.RawQuery = query.Encode()
	}
	return paginator.res, err
}

type JournalistInfo struct {
	ProfileURL *url.URL
	Name       string
	ImageName  string
	Gender     string
}

var bylineHtmlRegex = regexp.MustCompile(`<a.*?>[^<]+</a>`)

func JournalistInfoFromByline(byline string, webUrl *url.URL) []JournalistInfo {
	profiles := make([]JournalistInfo, 0)

	var updateProfilesFromMatch func(node *html.Node)
	updateProfilesFromMatch = func(node *html.Node) {
		if node.Type == html.ElementNode && node.DataAtom == atom.A {
			href := getAttrVal(node, "href")
			profile, err := url.Parse(href)
			if err != nil {
				fmt.Printf("unable to parse profile href %s: %s\n", href, err)
				return
			}
			// Url in byline HTML is relative.
			// Infer scheme and host from web url of the associated CAPI response.
			if webUrl != nil {
				profile.Scheme = webUrl.Scheme
				profile.Host = webUrl.Host
			}
			// Required for requests with this URL to execute successfully.
			if !strings.HasPrefix(profile.Path, "/") {
				profile.Path = fmt.Sprintf("/%s", profile.Path)
			}
			pathSegments := strings.Split(profile.Path, "/")
			profiles = append(profiles, JournalistInfo{
				ProfileURL: profile,
				// Know from inspection the last segment is the journalist's name.
				Name: pathSegments[len(pathSegments)-1],
			})
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			updateProfilesFromMatch(c)
		}
	}

	matches := bylineHtmlRegex.FindAllString(byline, -1)

	// e.g. Associated Press
	if len(matches) == 0 {
		profiles = append(profiles, JournalistInfo{Name: byline})
		return profiles
	}

	// Can be more than one match i.e. for articles written by multiple journalists
	for _, match := range matches {
		if node, err := html.Parse(strings.NewReader(match)); err != nil {
			fmt.Printf("unable to parse match %s as HTML: %s\n", match, err)
		} else {
			updateProfilesFromMatch(node)
		}
	}

	return profiles
}
