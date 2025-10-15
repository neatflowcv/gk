package domain

type File struct {
	path  string
	rel   string
	isDir bool
}

func NewFile(path string, rel string, isDir bool) *File {
	return &File{
		path:  path,
		rel:   rel,
		isDir: isDir,
	}
}

func (f *File) Path() string {
	return f.path
}

func (f *File) IsDir() bool {
	return f.isDir
}

func (f *File) Rel() string {
	return f.rel
}
