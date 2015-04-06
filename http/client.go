package http

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	iconv "github.com/djimenez/iconv-go"
	"github.com/gregjones/httpcache"
	"github.com/mcuadros/go-mgo-cache"
	"golang.org/x/net/html"
	"gopkg.in/mgo.v2"
)

type Client struct {
	http.Client
}

func NewCachedClient(session *mgo.Session) *Client {
	collection := session.DB("crawler").C("cache")

	transport := httpcache.NewTransport(mgocache.New(collection))
	transport.Transport = &responseModifier{}

	cli := &Client{}
	cli.Transport = transport

	return cli
}

func NewClient(cacheEnfornced bool) *Client {
	session, _ := mgo.Dial("mongodb://localhost")
	collection := session.DB("crawler").C("cache")

	transport := httpcache.NewTransport(mgocache.New(collection))
	transport.Transport = &responseModifier{}

	cli := &Client{}
	if cacheEnfornced {
		cli.Transport = transport
	}

	return cli
}

func (c *Client) DoJSON(req *http.Request, result interface{}) (*http.Response, error) {
	res, err := c.Do(req)
	if err != nil {
		return res, err
	}

	if res.StatusCode >= 400 {
		return res, nil
	}

	body, err := getResponseBodyReader(res)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	d := json.NewDecoder(body)
	if err := d.Decode(result); err != nil {
		return res, err
	}

	return res, nil
}

func (c *Client) DoHTML(req *http.Request) (*goquery.Document, *http.Response, error) {
	res, err := c.Do(req)
	if err != nil {
		return nil, res, err
	}

	doc, err := c.buildDocument(res)
	if err != nil {
		return nil, res, err
	}

	return doc, res, nil
}

func (c *Client) buildDocument(res *http.Response) (*goquery.Document, error) {
	body, err := getResponseBodyReader(res)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var reader io.Reader
	reader = body

	ct := strings.Split(res.Header.Get("Content-Type"), "=")
	if len(ct) == 2 && ct[1] != "UTF-8" {
		reader, err = iconv.NewReader(reader, ct[1], "utf-8")
		if err != nil {
			return nil, err
		}
	}

	root, err := html.Parse(reader)
	if err != nil {
		return nil, err
	}

	return goquery.NewDocumentFromNode(root), nil
}

func getResponseBodyReader(res *http.Response) (io.ReadCloser, error) {
	var reader io.ReadCloser
	var err error
	reader = res.Body

	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
	}

	return reader, nil
}

func NewRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,de;q=0.6,es;q=0.4")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.95 Safari/537.36")

	return req, nil
}

type responseModifier struct {
	Transport http.RoundTripper
}

func (t *responseModifier) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	resp.Header.Set("cache-control", "max-age=2592000")
	return resp, nil
}