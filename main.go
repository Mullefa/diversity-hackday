package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html/atom"

	"golang.org/x/net/html"
)

type capiSearchResponse struct {
	Response struct {
		Total      int `json:"total"`
		StartIndex int `json:"startIndex"`
		PageSize   int `json:"pageSize"`
		Results    []struct {
			WebURL string `json:"webUrl"`
			Fields struct {
				BylineHTML string `json:"bylineHtml"`
			}
		} `json:"results"`
	} `json:"response"`
}

// TODO: retries, status codes

func doJsonRequest(client *http.Client, request *http.Request, target interface{}) error {
	res, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&target); err != nil {
		return fmt.Errorf("unable to unmarshal body to CAPI search response: %w", err)
	}
	return nil
}

func doHtmlRequest(client *http.Client, request *http.Request) (*html.Node, error) {
	res, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer res.Body.Close()
	node, err := html.Parse(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse response body as HTML: %w", err)
	}
	return node, nil
}

func getAttrVal(node *html.Node, key string) string {
	for _, a := range node.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func findAuthorsProfiles(root *html.Node, parent *url.URL) []*url.URL {
	profiles := make([]*url.URL, 0)
	// TODO: consider early break
	var search func(node *html.Node)
	search = func(node *html.Node) {
		if node.Type == html.ElementNode && node.DataAtom == atom.A {
			if getAttrVal(node, "rel") == "author" {
				href := getAttrVal(node, "href")
				if u, err := url.Parse(href); err != nil {
					fmt.Printf("unable to parse profile href %s: %s\n", href, err)
				} else {
					if u.Scheme == "" && parent != nil {
						u.Scheme = parent.Scheme
					}
					if u.Host == "" && parent != nil {
						u.Host = parent.Host
					}
					profiles = append(profiles, u)
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			search(c)
		}
	}
	search(root)
	return profiles
}

func main() {
	client := &http.Client{}

	apiKey := os.Getenv("CAPI_API_KEY")
	if apiKey == "" {
		fmt.Println("please specify API key via environment variable CAPI_API_KEY")
		return
	}

	u, err := url.Parse("https://content.guardianapis.com/search?use-date=first-publication&from-date=2020-01-01&to-date=2020-01-01&page-size=10&show-fields=byline")
	if err != nil {
		fmt.Printf("error parsing url: %s\n", err)
		return
	}

	q := u.Query()
	q.Add("api-key", apiKey)
	u.RawQuery = q.Encode()

	request := &http.Request{URL: u}

	var response capiSearchResponse
	if err := doJsonRequest(client, request, &response); err != nil {
		fmt.Printf("unable to execute JSON request: %s\n", err)
	}

	for _, result := range response.Response.Results {
		u, err := url.Parse(result.WebURL)
		if err != nil {
			fmt.Printf("unable to parse raw url %s\n", result.WebURL)
			continue
		}
		request := &http.Request{URL: u}
		root, err := doHtmlRequest(client, request)
		if err != nil {
			fmt.Printf("unable to execute request for url %s: %s\n", u, err)
		}
		profiles := findAuthorsProfiles(root, u)
		if len(profiles) == 0 {
			fmt.Printf("no profiles found for url %s\n", result.WebURL)
		} else {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("profiles for for url %s:\n", result.WebURL))
			for _, profile := range profiles {
				b.WriteString(fmt.Sprintf("- %s\n", profile))
			}
			fmt.Print(b.String())
		}
	}
}
