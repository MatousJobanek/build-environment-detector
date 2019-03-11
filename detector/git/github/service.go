package github

import (
	"context"
	"net/http"
	"net/url"

	"github.com/MatousJobanek/build-environment-detector/detector/git"
	gogh "github.com/google/go-github/github"
)

const (
	githubHost   = "github.com"
	githubFlavor = "github"
)

type GitService struct {
	gitSource  *git.Source
	client     *gogh.Client
	repository git.Repository
}

func NewGitServiceIfMatches() git.ServiceCreator {
	return func(gitSource *git.Source) (git.Service, error) {
		if gitSource.Secret.SecretType() == git.SshKeyType {
			return nil, nil
		}
		url, err := url.Parse(gitSource.URL)
		if err != nil {
			return nil, err
		}

		if url.Host == githubHost || gitSource.Flavor == githubFlavor {
			return newGhClient(gitSource, url)
		}
		return nil, nil
	}
}

func newGhClient(gitSource *git.Source, url *url.URL) (*GitService, error) {
	repository, err := git.NewRepository(gitSource, url)
	if err != nil {
		return nil, err
	}
	baseClient := gitSource.Secret.Client()
	if gitSource.Secret.SecretType() == git.UsernamePasswordType {
		username, password := git.ParseUsernameAndPassword(gitSource.Secret.Content())
		baseClient.Transport = &gogh.BasicAuthTransport{Username: username, Password: password}
	}
	client := gogh.NewClient(baseClient)
	if err = getBranchRequestErrors(context.Background(), client, repository); err != nil {
		return nil, err
	}
	return &GitService{
		gitSource:  gitSource,
		client:     client,
		repository: repository,
	}, nil
}

func (s *GitService) Exists(filePath string) bool {
	_, _, resp, err := s.client.Repositories.GetContents(
		context.Background(),
		s.repository.Owner,
		s.repository.Repository,
		filePath,
		&gogh.RepositoryContentGetOptions{Ref: s.repository.Branch})

	return err == nil && resp != nil && resp.StatusCode == http.StatusOK
}

func (s *GitService) GetLanguageList() ([]string, error) {
	languages, _, err := s.client.Repositories.ListLanguages(
		context.Background(),
		s.repository.Owner,
		s.repository.Repository)

	if err != nil {
		return nil, err
	}

	return git.GetSortedLanguages(languages), nil
}

func getBranchRequestErrors(ctx context.Context, client *gogh.Client, repository git.Repository) error {
	_, _, err := client.Repositories.GetBranch(ctx, repository.Owner, repository.Repository, repository.Branch)
	return err
}
