package profiles

type GlobalConfig struct {
	Ignore []string `json:"ignore"`
}

type DetectCriteria struct {
	Files      []string `json:"files"`
	Folders    []string `json:"folders"`
	Extensions []string `json:"extensions"`
}

type Profile struct {
	Name       string         `json:"name"`
	Priority   int            `json:"priority"`
	Detect     DetectCriteria `json:"detect"`
	Extensions []string       `json:"extensions"`
	Ignore     []string       `json:"ignore"`
	Generated  []string       `json:"generated"`
}

type Config struct {
	Global   GlobalConfig `json:"global"`
	Profiles []Profile    `json:"profiles"`
}
