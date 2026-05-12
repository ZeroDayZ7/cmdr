package annotate

func GetCommentStyle(ext string, cfg Config) (string, bool) {
	style, ok := cfg.ProfilesConfig.CommentStyles[ext]
	return style, ok
}
