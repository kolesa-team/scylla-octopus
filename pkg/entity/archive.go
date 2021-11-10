package entity

// Archive method and options for compress
type Archive struct {
	Method  string  `yaml:"method"`
	Options Options `yaml:"options"`
}

// Options options for compress. compression level and number of threads used for compression
type Options struct {
	Compression string `yaml:"compression"`
	Threads     string `yaml:"threads"`
}
