package annotate_origin

func GetCommentStyle(ext string, cfg Config) (string, bool) {
	style, ok := cfg.ProfilesConfig.CommentStyles[ext]
	return style, ok
}
