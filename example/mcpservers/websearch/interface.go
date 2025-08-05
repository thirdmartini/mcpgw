package main

type SearchProvider interface {
	Search(query string, results int) (string, error)
	News(query string, results int) (string, error)
}

type NewsResult struct {
	Title       string
	Description string
	Summary     string
	//Link        string ( for this to work need to add support for links in ui )
}

type SearchResult struct {
	Title       string
	Description string
	Summary     string
	//Link        string ( for this to work need to add support for links in ui )
}
