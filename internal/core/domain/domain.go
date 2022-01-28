package domain

type ActionFunc func() error

type Commands struct {
	Action   ActionFunc
	IsLeaf   bool
	Children map[string]*Commands
}

type Server struct {
	APIKey string
	*Commands
}

type Config struct {
	WOL WOL `yaml:"wol"`
}

type WOL struct {
	Commands   []string `yaml:"commands"`
	MACAddress string   `yaml:"macAddress"`
}
