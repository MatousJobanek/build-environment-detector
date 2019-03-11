package generic

import (
	"fmt"

	"github.com/MatousJobanek/build-environment-detector/detector/git"
	"github.com/Sirupsen/logrus"
	"gopkg.in/src-d/enry.v1"
	"gopkg.in/src-d/go-billy.v4/memfs"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type GitService struct {
	repository *gogit.Repository
	tree       *object.Tree
	fileNames  []string
}

func NewGitService(gitSource *git.Source) (git.Service, error) {
	storage := memory.NewStorage()
	branch := git.Master
	if gitSource.Ref != "" {
		branch = gitSource.Ref
	}
	refSpec := fmt.Sprintf("+refs/heads/%[1]s:refs/remotes/origin/%[1]s", branch)
	repository, err := gogit.Init(storage, memfs.New())
	repository.CreateRemote(&config.RemoteConfig{
		Name:  "origin",
		URLs:  []string{gitSource.URL},
		Fetch: []config.RefSpec{config.RefSpec(refSpec)},
	})

	authMethod, err := gitSource.Secret.GitAuthMethod()
	if err != nil {
		return nil, err
	}

	err = repository.Fetch(&gogit.FetchOptions{
		Auth:       authMethod,
		Depth:      1,
		Tags:       gogit.NoTags,
		RemoteName: "origin",
	})
	if err != nil {
		return nil, err
	}

	commitIter, err := repository.CommitObjects()
	commitToList, err := commitIter.Next()

	commit, err := repository.CommitObject(commitToList.Hash)
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	service := &GitService{
		repository: repository,
		tree:       tree,
	}
	tree.Files().ForEach(func(f *object.File) error {
		service.fileNames = append(service.fileNames, f.Name)
		return nil
	})

	return service, nil
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
	languagesCounts := map[string]int{}
	s.tree.Files().ForEach(func(f *object.File) error {
		language, safe := enry.GetLanguageByExtension(f.Name)
		if safe {
			languagesCounts[language]++
		} else {
			language, safe := enry.GetLanguageByFilename(f.Name)
			if safe {
				languagesCounts[language]++
			} else {
				content, err := f.Contents()
				if err != nil {
					logrus.Warn(err)
				} else {
					language, safe = enry.GetLanguageByContent(f.Name, []byte(content))
				}
				if safe {
					languagesCounts[language]++
				}
			}
		}
		return nil
	})

	return git.GetSortedLanguages(languagesCounts), nil
}
