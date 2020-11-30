package internal

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var NoProfileImage = errors.New("no profile image")

func ImageUrlFromProfileUrl(client *http.Client, profile *url.URL) (*url.URL, error) {
	request := &http.Request{URL: profile}
	root, err := doHtmlRequest(client, request)
	if err != nil {
		return nil, fmt.Errorf("unable to get HTML: %w\n", err)
	}

	var searchErr error
	var imageUrl *url.URL

	// TODO: optimise
	var getImageUrl func(node *html.Node)
	getImageUrl = func(node *html.Node) {
		if node.Type == html.ElementNode && node.DataAtom == atom.Img && getAttrVal(node, "class") == "index-page-header__image" {
			href := getAttrVal(node, "src")
			if u, err := url.Parse(href); err != nil {
				searchErr = fmt.Errorf("unable to parse profile image src %s: %w", href, err)
			} else {
				imageUrl = u
			}
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			getImageUrl(c)
		}
	}
	getImageUrl(root)

	if searchErr == nil && imageUrl == nil {
		searchErr = NoProfileImage
	}

	return imageUrl, searchErr
}
