package git

type Source struct {
	// URL of the git repo
	URL string

	// Ref is a git reference. Optional. "Master" is used by default.
	Ref string

	// ContextDir is a path to subfolder in the repo. Optional.
	ContextDir string

	// HttpProxy is optional.
	HttpProxy string

	// HttpsProxy is optional.
	HttpsProxy string

	// NoProxy can be used to specify domains for which no proxying should be performed. Optional.
	NoProxy string

	// Secret refers to the credentials to access the git repo. Optional.
	Secret Secret

	// Flavor of the git provider like github, gitlab, bitbucket, generic, etc. Optional.
	Flavor string
}
