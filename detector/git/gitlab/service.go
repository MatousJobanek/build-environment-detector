package gitlab

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/MatousJobanek/build-environment-detector/detector/git"
	gogl "github.com/xanzy/go-gitlab"
)

const (
	gitlabHost   = "gitlab.com"
	gitlabFlavor = "gitlab"
)

type GitService struct {
	gitSource  *git.Source
	client     *gogl.Client
	repository git.Repository
	fileNames  []string
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

		if url.Host == gitlabHost || gitSource.Flavor == gitlabFlavor {
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

	client := gogl.NewOAuthClient(gitSource.Secret.Client(), gitSource.Secret.Content())
	if gitSource.Secret.SecretType() == git.UsernamePasswordType {
		username, password := git.ParseUsernameAndPassword(gitSource.Secret.Content())
		shortUrl := *url
		shortUrl.Path = ""
		client, err = gogl.NewBasicAuthClient(gitSource.Secret.Client(), shortUrl.String(), username, password)
		if err != nil {
			return nil, err
		}
	}
	fileNames, err := getListOfFiles(client, repository)
	if err != nil {
		return nil, err
	}

	return &GitService{
		gitSource:  gitSource,
		client:     client,
		repository: repository,
		fileNames:  fileNames,
	}, nil
}

func getListOfFiles(client *gogl.Client, repository git.Repository) ([]string, error) {
	tree, _, err := client.Repositories.ListTree(
		getPid(repository),
		&gogl.ListTreeOptions{})
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, entry := range tree {
		filenames = append(filenames, entry.Path)
	}
	return filenames, nil
}

func (s *GitService) Exists(filePath string) bool {
	for _, fileName := range s.fileNames {
		if fileName == filePath {
			return true
		}
	}
	return false
}

func (s *GitService) GetLanguageList() ([]string, error) {
	languages, _, err := s.client.Projects.GetProjectLanguages(getPid(s.repository))
	if err != nil {
		return nil, err
	}
	var contentSizes []float64
	reversedMap := map[float64]string{}
	for lang, size := range *languages {
		reversedMap[float64(size)] = lang
		contentSizes = append(contentSizes, float64(size))
	}
	sort.Float64s(contentSizes)

	var sortedLangs []string
	for i := len(contentSizes) - 1; i >= 0; i-- {
		sortedLangs = append(sortedLangs, reversedMap[contentSizes[i]])
	}
	return sortedLangs, nil
}

func getPid(repository git.Repository) string {
	return fmt.Sprintf("%s/%s", repository.Owner, repository.Repository)
}
