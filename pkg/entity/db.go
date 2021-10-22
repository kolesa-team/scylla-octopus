package entity

// NodeBinaries paths to scylladb executables on database node
type NodeBinaries struct {
	Cqlsh    string
	Nodetool string
}

// Credentials scylladb credentials
type Credentials struct {
	User     string
	Password string
}
