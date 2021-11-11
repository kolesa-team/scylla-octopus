package entity

// Archive method and options for compress
type Archive struct {
	Method         string         `yaml:"method"`
	ArchiveOptions ArchiveOptions `yaml:"options"`
}

// ArchiveOptions options for compress. compression level and number of threads used for compression
type ArchiveOptions struct {
	Compression string `yaml:"compression"`
	Threads     string `yaml:"threads"`
}
