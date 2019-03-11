package git

import (
	"sort"
)

type Service interface {
	Exists(filePath string) bool
	GetLanguageList() ([]string, error)
}

type ServiceCreator func(gitSource *Source) (Service, error)

func NewService(gitSource *Source, serviceCreators []ServiceCreator) (Service, error) {

	for _, creator := range serviceCreators {
		detector, err := creator(gitSource)
		if err != nil {
			return nil, err
		}
		if detector != nil {
			return detector, nil
		}
	}

	return nil, nil
}

func GetSortedLanguages(langsWithSizes map[string]int) []string {
	var contentSizes []int
	reversedMap := map[int]string{}
	for lang, size := range langsWithSizes {
		reversedMap[size] = lang
		contentSizes = append(contentSizes, size)
	}
	sort.Ints(contentSizes)

	var sortedLangs []string
	for i := len(contentSizes) - 1; i >= 0; i-- {
		sortedLangs = append(sortedLangs, reversedMap[contentSizes[i]])
	}
	return sortedLangs
}
