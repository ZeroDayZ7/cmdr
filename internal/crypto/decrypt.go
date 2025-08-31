package crypto

func DecryptFile(inputPath string, key []byte, outPath string) error {
	return DecryptFileAES(inputPath, key, outPath)
}

func DecryptFolder(inputPath string, key []byte, outDir string) error {
	return DecryptFolderAES(inputPath, key, outDir)
}
