package git

import (
	"errors"
	"net/url"
	"strings"
)

const (
	slash  = "/"
	Master = "master"
)

type Repository struct {
	Owner      string
	Repository string
	Branch     string
	Token      string
}

func NewRepository(gitSource *Source, url *url.URL) (Repository, error) {
	var repo Repository

	branch := Master
	urlSegments := strings.Split(url.Path, slash)

	if len(urlSegments) < 3 {
		return repo, errors.New("url is invalid")
	}

	if gitSource.Ref != "" {
		branch = gitSource.Ref
	}

	repo = Repository{
		Owner:      urlSegments[1],
		Repository: urlSegments[2],
		Branch:     branch,
	}

	return repo, nil
}
