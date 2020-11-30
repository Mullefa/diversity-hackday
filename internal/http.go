package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

func doJsonRequest(client *http.Client, request *http.Request, target interface{}) error {
	res, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return fmt.Errorf("unable to unmarshal body to res search res: %w", err)
	}
	return nil
}

func doHtmlRequest(client *http.Client, request *http.Request) (*html.Node, error) {
	doRequest := func() (int, *html.Node, error) {
		res, err := client.Do(request)
		if err != nil {
			return -1, nil, fmt.Errorf("error executing request: %w", err)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			return res.StatusCode, nil, fmt.Errorf("%d GET %s", res.StatusCode, request.URL)
		}

		node, err := html.Parse(res.Body)
		if err != nil {
			return 200, nil, fmt.Errorf("unable to parse res body as HTML: %w", err)
		}

		return 200, node, nil
	}

	for retries := 0; ; retries++ {
		status, node, err := doRequest()
		// Not an attempt for a general retry mechanism.
		// Written in response to Dotcom returning 429 i.e. too many requests.
		// Mitigate against this with exponential back off.
		if status == 429 && retries < 8 {
			exp := int(math.Pow(2, float64(retries)))
			time.Sleep(100 * time.Duration(exp) * time.Millisecond)
			continue
		}
		return node, err
	}
}

type ImageType int

const (
	PNG ImageType = iota
	JPEG
	NA
)

func (it ImageType) Extension() string {
	switch it {
	case PNG:
		return "png"
	case JPEG:
		return "jpeg"
	default:
		return ""
	}
}

func DownloadImage(client *http.Client, url *url.URL) (ImageType, []byte, error) {
	request := &http.Request{URL: url}
	res, err := client.Do(request)
	if err != nil {
		return NA, nil, fmt.Errorf("unable to download image: %w", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return NA, nil, fmt.Errorf("unable to read rsponse body: %w", err)
	}

	contentType := http.DetectContentType(body)
	switch contentType {
	case "image/jpeg":
		return JPEG, body, nil
	case "image/png":
		return PNG, body, nil
	default:
		return NA, nil, fmt.Errorf("don't know how to handle image with content type %s", contentType)
	}
}
