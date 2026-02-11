package commands

type ymlFile struct {
	Filename  string `yaml:"filename"`
	Timestamp int64  `yaml:"timestamp"`
	Index     int    `yaml:"index"`
	Total     int    `yaml:"total"`
	Minimum   int    `yaml:"minimum"`
	Keypart   string `yaml:"keypart"`
	Payload   string `yaml:"payload"`
}

