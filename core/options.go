package core

// Options defines configuration parameters for pagination behavior.
type Options struct {
	// PageSize determines how many rows to fetch per page.
	// Default: 100 if not specified.
	PageSize int
}
