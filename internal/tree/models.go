package tree

type Node struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Children []Node `json:"children,omitempty"`
}

type TreeOptions struct {
	Path        string
	ExcludeList []string
	Format      string
	MaxDepth    int
}
