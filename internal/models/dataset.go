package models

// Dataset represents a single data source loaded into memory.
type Dataset struct {
	// Name is the endpoint name (without a leading slash).
	Name string

	// Path is the absolute path of the source file.
	Path string

	// Data is the parsed root value.
	Data any

	// Format is the source format (file extension without the dot).
	Format string
}
