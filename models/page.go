package models

// Many of the APIs that are called will page results
// in the case that the called wants to handle paging this is
// a convenient object that is used in the Cloudy API.
type Page struct {
	// The number of results to skip
	Skip int

	// The size of the page
	PageSize int

	// Next Page Token
	NextPageToken interface{}
}
