package commands

import (
	"time"

	"github.com/src-d/rovers/readers"
	"gop.kg/src-d/domain@v6/container"
	"gop.kg/src-d/domain@v6/models/social"

	"github.com/mcuadros/go-github/github"
	"gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/src-d/storable.v1"
)

type CmdGitHubAPIRepos struct {
	github    *readers.GithubAPI
	repoStore *social.GithubRepositoryStore
	userStore *social.GithubUserStore
}

func (c *CmdGitHubAPIRepos) Execute(args []string) error {
	c.github = readers.NewGithubAPI()
	c.repoStore = container.GetDomainModelsSocialGithubRepositoryStore()
	c.userStore = container.GetDomainModelsSocialGithubUserStore()

	start := time.Now()
	since := c.getSince()
	for {
		log15.Info("Requesting repositories...", "since", since)

		repos, resp, err := c.github.GetAllRepositories(since)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			log15.Info("No more repos. Stopping crawl...")
			break
		}

		c.save(repos)

		if resp.NextPage == 0 && resp.NextPage == since {
			break
		}

		since = resp.NextPage
	}

	log15.Info("Done", "elapsed", time.Since(start))
	return nil
}

func (c *CmdGitHubAPIRepos) getSince() int {
	q := c.repoStore.Query()

	q.Sort(storable.Sort{{social.Schema.GithubRepository.GithubID, storable.Desc}})
	repo, err := c.repoStore.FindOne(q)
	if err != nil {
		return 0
	}

	return repo.GithubID
}

func (c *CmdGitHubAPIRepos) getRepositories(since int) (
	repos []github.Repository, resp *github.Response, err error,
) {
	start := time.Now()
	repos, resp, err = c.github.GetAllRepositories(since)
	if err != nil {
		log15.Error("GetAllRepositories failed",
			"since", since,
			"error", err,
		)
		return
	}

	elapsed := time.Since(start)
	microseconds := float64(elapsed) / float64(time.Microsecond)
	return
}

func (c *CmdGitHubAPIRepos) save(repos []github.Repository) {
	for _, repo := range repos {
		doc := c.createNewDocument(repo)
		if _, err := c.repoStore.Save(doc); err != nil {
			log15.Error("Repository save failed",
				"repo", doc.FullName,
				"error", err,
			)
		}
	}

	numRepos := len(repos)
	log15.Info("Repositories saved", "num_repos", numRepos)
}

func (c *CmdGitHubAPIRepos) createNewDocument(repo github.Repository) *social.GithubRepository {
	doc := c.repoStore.New()
	doc.Owner = c.userStore.New()
	processGithubRepository(doc, repo)
	return doc
}

func processGithubRepository(doc *social.GithubRepository, repo github.Repository) {
	if repo.ID != nil {
		doc.GithubID = *repo.ID
	}
	if repo.Name != nil {
		doc.Name = *repo.Name
	}
	if repo.FullName != nil {
		doc.FullName = *repo.FullName
	}
	if repo.Description != nil {
		doc.Description = *repo.Description
	}
	if repo.Owner != nil {
		processGithubUser(doc.Owner, *repo.Owner)
		doc.Owner.SetId(bson.NewObjectId())
	}
	if repo.Fork != nil {
		doc.Fork = *repo.Fork
	}
}
