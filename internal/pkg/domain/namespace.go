package domain

func IsNamespace(file *File) bool {
	return file.IsDir()
}
