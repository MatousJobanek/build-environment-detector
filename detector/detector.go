package detector

import (
	"sync"

	"github.com/MatousJobanek/build-environment-detector/detector/environment"
	"github.com/MatousJobanek/build-environment-detector/detector/git"
	"github.com/MatousJobanek/build-environment-detector/detector/git/generic"
	"github.com/MatousJobanek/build-environment-detector/detector/git/github"
	"github.com/MatousJobanek/build-environment-detector/detector/git/gitlab"
)

var gitServiceCreators = []git.ServiceCreator{
	github.NewGitServiceIfMatches(),
	gitlab.NewGitServiceIfMatches(),
}

func DetectBuildEnvironments(gitSource *git.Source) (*environment.BuildEnvStats, error) {

	service, err := git.NewService(gitSource, gitServiceCreators)
	if err != nil {
		return nil, err
	}
	if service == nil {
		service, err = generic.NewGitService(gitSource)
		if err != nil {
			return nil, err
		}
	}
	environments := detectBuildEnvironments(service)
	languageList, err := service.GetLanguageList()
	if err != nil {
		return nil, err
	}

	return &environment.BuildEnvStats{
		SortedLanguages:    languageList,
		DetectedBuildTools: environments,
	}, nil
}

func detectBuildEnvironments(service git.Service) []*environment.DetectedBuildTool {
	detectedBuildTools := make(chan *environment.DetectedBuildTool, len(environment.BuildTools))
	var wg sync.WaitGroup
	wg.Add(len(environment.BuildTools))

	for _, tool := range environment.BuildTools {
		go func(buildTool environment.BuildTool) {
			defer wg.Done()
			detectedFiles := detectBuildToolFiles(service, buildTool)
			if len(detectedFiles) > 0 {
				detectedBuildTools <- environment.NewDetectedBuildTool(buildTool.Language, buildTool.Name, detectedFiles)
			}
		}(tool)
	}

	wg.Wait()
	close(detectedBuildTools)
	var result []*environment.DetectedBuildTool
	for detectedBuildTool := range detectedBuildTools {
		if detectedBuildTool != nil {
			result = append(result, detectedBuildTool)
		}
	}
	return result
}

func detectBuildToolFiles(service git.Service, buildTool environment.BuildTool) []string {
	detectedFiles := make(chan string, len(buildTool.ExpectedFiles))
	var wg sync.WaitGroup
	wg.Add(len(buildTool.ExpectedFiles))

	for _, file := range buildTool.ExpectedFiles {
		go func(buildToolFile string) {
			defer wg.Done()
			if service.Exists(buildToolFile) {
				detectedFiles <- buildToolFile
			}
		}(file)
	}

	wg.Wait()
	close(detectedFiles)
	var result []string
	for detectedFile := range detectedFiles {
		if detectedFile != "" {
			result = append(result, detectedFile)
		}
	}

	return result
}
