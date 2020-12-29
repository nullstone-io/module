package config

func ParseDir(dir string) (*InternalTfConfig, error) {
	files, err := ReadDir(dir)
	if err != nil {
		return nil, err
	}
	return ParseFiles(files)
}
