package github

import (
	"context"
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
	filenames  []string
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
			return newGhService(gitSource, url)
		}
		return nil, nil
	}
}

func newGhService(gitSource *git.Source, url *url.URL) (*GitService, error) {
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
	listOfFiles, err := getListOfFiles(client, repository)
	if err != nil {
		return nil, err
	}

	return &GitService{
		gitSource:  gitSource,
		client:     client,
		repository: repository,
		filenames:  listOfFiles,
	}, nil
}

func getListOfFiles(client *gogh.Client, repository git.Repository) ([]string, error) {
	tree, _, err := client.Git.GetTree(
		context.Background(),
		repository.Owner,
		repository.Repository,
		repository.Branch,
		false)
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, entry := range tree.Entries {
		filenames = append(filenames, *entry.Path)
	}
	return filenames, nil
}

func (s *GitService) Exists(filePath string) bool {
	for _, file := range s.filenames {
		if filePath == file {
			return true
		}
	}
	return false
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
