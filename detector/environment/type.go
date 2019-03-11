package environment

type BuildEnvStats struct {
	SortedLanguages    []string
	DetectedBuildTools []*DetectedBuildTool
}

type DetectedBuildTool struct {
	language      string
	name          string
	detectedFiles []string
}

type BuildTool struct {
	Language      string
	Name          string
	ExpectedFiles []string
}

func NewDetectedBuildTool(language string, name string, detectedFiles []string) *DetectedBuildTool {
	return &DetectedBuildTool{
		language:      language,
		name:          name,
		detectedFiles: detectedFiles,
	}
}

var BuildTools = []BuildTool{Maven, Gradle, Golang, Ruby}

var Maven = BuildTool{
	Name:          "Maven",
	Language:      "java",
	ExpectedFiles: []string{"pom.xml"},
}

var Gradle = BuildTool{
	Name:          "Gradle",
	Language:      "java",
	ExpectedFiles: []string{"gradle"},
}

var Golang = BuildTool{
	Name:          "Golang",
	Language:      "go",
	ExpectedFiles: []string{"main.go", "Gopkg.toml", "glide.yaml"},
}

var Ruby = BuildTool{
	Name:          "Ruby",
	Language:      "ruby",
	ExpectedFiles: []string{"Gemfile"},
}
