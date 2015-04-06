package commands

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/tyba/opensource-search/domain/models/social"
	"github.com/tyba/opensource-search/sources/social/http"
	"github.com/tyba/opensource-search/sources/social/readers"

	"gopkgs.com/unidecode.v1"
)

type CmdPlayground struct {
	FullName string `short:"f" long:"fullname" description:"fullname"`
	Tags     string `short:"t" long:"tags" description:"tags"`
}

func (t *CmdPlayground) Execute(args []string) error {
	c := NewCrawler()
	g, _ := c.SearchLinkedIn(t.FullName, strings.Split(t.Tags, " ")...)

	j, _ := json.MarshalIndent(g, "", "    ")
	fmt.Println(string(j))

	return nil
}

const linkedInSearch = "linkedin %s %s"
const githubSearch = "linkedin %s %s"

var justCharsRegexp = regexp.MustCompile("[^a-z -]")
var smallWordsRegexp = regexp.MustCompile("\\b\\w{1,2}\\b\\s?")

type Crawler struct {
	client *http.Client
}

func NewCrawler() *Crawler {
	return &Crawler{http.NewClient(true)}
}

func (c *Crawler) SearchLinkedIn(fullname string, tags ...string) (*social.LinkedInProfile, error) {
	q := fmt.Sprintf(linkedInSearch, fullname, strings.Join(tags, " "))
	r, err := readers.NewGoogleSearchReader(c.client).SearchByQuery(q)
	if err != nil {
		return nil, err
	}

	url := c.findLinkedInBestMatchURL(fullname, r)
	if url == "" {
		return nil, nil
	}

	return readers.NewLinkedInReader(c.client).GetProfileByURL(url)
}

func (c *Crawler) findLinkedInBestMatchURL(fullname string, search *readers.GoogleSearchResult) string {
	n := c.normalize(fullname)
	for _, r := range search.FindByHost("linkedin.com") {
		l := strings.Split(r.Name, "|")
		t := c.normalize(l[0])

		if strings.HasPrefix(t, n) && !strings.Contains(t, "profiles") {
			return r.Link
		}
	}

	return ""
}

func (c *Crawler) SearchGithub(fullname string, tags ...string) (*social.GithubProfile, error) {
	q := fmt.Sprintf(githubSearch, fullname, strings.Join(tags, " "))
	r, err := readers.NewGoogleSearchReader(c.client).SearchByQuery(q)
	if err != nil {
		return nil, err
	}

	urls := r.FindByHost("github.com")
	if len(urls) == 0 {
		return nil, nil
	}

	return readers.NewGithubReader(c.client).GetProfileByURL(urls[0].Link)
}

func (c *Crawler) normalize(name string) string {
	u := unidecode.Unidecode(name)
	u = strings.ToLower(u)
	u = justCharsRegexp.ReplaceAllString(u, "")
	u = smallWordsRegexp.ReplaceAllString(u, "")

	return u
}