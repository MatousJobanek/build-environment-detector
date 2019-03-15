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
	ExpectedFiles: []string{"*gradle*"},
}

var Golang = BuildTool{
	Name:          "Golang",
	Language:      "go",
	ExpectedFiles: []string{"main.go", "Gopkg.toml", "glide.yaml"},
}

var Ruby = BuildTool{
	Name:          "Ruby",
	Language:      "ruby",
	ExpectedFiles: []string{"Gemfile", "Rakefile", "config.ru"},
}

var NodeJS = BuildTool{
	Name:          "NodeJS",
	Language:      "javascript",
	ExpectedFiles: []string{"app.json", "package.json", "gulpfile.js", "Gruntfile.js"},
}

var PHP = BuildTool{
	Name:          "PHP",
	Language:      "php",
	ExpectedFiles: []string{"index.php", "composer.json"},
}

var Python = BuildTool{
	Name:          "Python",
	Language:      "python",
	ExpectedFiles: []string{"requirements.txt", "setup.py"},
}

var Perl = BuildTool{
	Name:          "Perl",
	Language:      "perl",
	ExpectedFiles: []string{"index.pl", "cpanfile"},
}

var Dotnet = BuildTool{
	Name:          "Dotnet",
	Language:      "C#",
	ExpectedFiles: []string{"project.json", "*.csproj"},
}
