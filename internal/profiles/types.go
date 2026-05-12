package profiles

type GlobalConfig struct {
	Ignore            []string `json:"ignore"`
	IgnoredExtensions []string `json:"ignored_extensions"`
}

type DetectCriteria struct {
	Files      []string `json:"files"`
	Folders    []string `json:"folders"`
	Extensions []string `json:"extensions"`
}

type RegionPattern struct {
	Regex string `json:"regex"`
	Style string `json:"style"`
}

type Profile struct {
	Name           string          `json:"name"`
	Priority       int             `json:"priority"`
	Detect         DetectCriteria  `json:"detect"`
	Extensions     []string        `json:"extensions"`
	Ignore         []string        `json:"ignore"`
	Generated      []string        `json:"generated"`
	RegionPatterns []RegionPattern `json:"region_patterns"`
}

type Config struct {
	Global        GlobalConfig      `json:"global"`
	Profiles      []Profile         `json:"profiles"`
	CommentStyles map[string]string `json:"comment_styles"`
	IgnoredExts   []string          `json:"ignored_exts"`
}
