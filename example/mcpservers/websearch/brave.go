package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type BraveNewsResponse struct {
	/* We don't care about this
	Type  string `json:"type"`
	Query struct {
		Original          string `json:"original"`
		SpellcheckOff     bool   `json:"spellcheck_off"`
		ShowStrictWarning bool   `json:"show_strict_warning"`
	} `json:"query"`
	*/
	Results []struct {
		Type        string `json:"type"`
		Title       string `json:"title"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Age         string `json:"age"`
		PageAge     string `json:"page_age"`
		MetaURL     struct {
			Scheme   string `json:"scheme"`
			Netloc   string `json:"netloc"`
			Hostname string `json:"hostname"`
			Favicon  string `json:"favicon"`
			Path     string `json:"path"`
		} `json:"meta_url"`
		Thumbnail struct {
			Src string `json:"src"`
		} `json:"thumbnail,omitempty"`
		ExtraSnippets []string `json:"extra_snippets,omitempty"`
	} `json:"results"`
}

type BraveSearchResponse struct {
	/* We don;t care about this
	Query struct {
		Original             string `json:"original"`
		ShowStrictWarning    bool   `json:"show_strict_warning"`
		IsNavigational       bool   `json:"is_navigational"`
		IsNewsBreaking       bool   `json:"is_news_breaking"`
		SpellcheckOff        bool   `json:"spellcheck_off"`
		Country              string `json:"country"`
		BadResults           bool   `json:"bad_results"`
		ShouldFallback       bool   `json:"should_fallback"`
		PostalCode           string `json:"postal_code"`
		City                 string `json:"city"`
		HeaderCountry        string `json:"header_country"`
		MoreResultsAvailable bool   `json:"more_results_available"`
		State                string `json:"state"`
	} `json:"query"`
	Mixed struct {
		Type string `json:"type"`
		Main []struct {
			Type  string `json:"type"`
			Index int    `json:"index"`
			All   bool   `json:"all"`
		} `json:"main"`
		Top  []any `json:"top"`
		Side []any `json:"side"`
	} `json:"mixed"`
	Type string `json:"type"`*/
	Web struct {
		Type    string `json:"type"`
		Results []struct {
			Title         string `json:"title"`
			URL           string `json:"url"`
			IsSourceLocal bool   `json:"is_source_local"`
			IsSourceBoth  bool   `json:"is_source_both"`
			Description   string `json:"description"`
			PageAge       string `json:"page_age"`
			Profile       struct {
				Name     string `json:"name"`
				URL      string `json:"url"`
				LongName string `json:"long_name"`
				Img      string `json:"img"`
			} `json:"profile"`
			Language       string `json:"language"`
			FamilyFriendly bool   `json:"family_friendly"`
			Type           string `json:"type"`
			Subtype        string `json:"subtype"`
			IsLive         bool   `json:"is_live"`
			MetaURL        struct {
				Scheme   string `json:"scheme"`
				Netloc   string `json:"netloc"`
				Hostname string `json:"hostname"`
				Favicon  string `json:"favicon"`
				Path     string `json:"path"`
			} `json:"meta_url"`
			Thumbnail struct {
				Src      string `json:"src"`
				Original string `json:"original"`
				Logo     bool   `json:"logo"`
			} `json:"thumbnail"`
			Age           string   `json:"age"`
			ExtraSnippets []string `json:"extra_snippets"` // snippets are only available when using the AI Token
		} `json:"results"`
		FamilyFriendly bool `json:"family_friendly"`
	} `json:"web"`
}

func (b *BraveSearch) Search(query string, results int) (string, error) {
	u, _ := url.Parse("https://api.search.brave.com/res/v1/web/search")

	q := u.Query()
	q.Set("q", query)
	q.Set("count", fmt.Sprintf("%d", results))
	q.Set("country", "us")
	q.Set("search_lang", "en")
	q.Set("summary", "1")
	q.Set("extra_snippets", "1")
	q.Set("result_filter", "web")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("X-Subscription-Token", b.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response with status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	searchResp := &BraveSearchResponse{}
	if err = json.Unmarshal(body, searchResp); err != nil {
		return "", err
	}

	searchResults := make([]SearchResult, 0, len(searchResp.Web.Results))
	for _, web := range searchResp.Web.Results {
		searchResults = append(searchResults, SearchResult{
			Title:       web.Title,
			Summary:     web.Description,
			Description: strings.Join(web.ExtraSnippets, "\n"),
			//Link:        web.URL,
		})
	}

	data, err := json.Marshal(searchResults)
	return string(data), err
}

type BraveSearch struct {
	apiKey string
}

func (b *BraveSearch) News(query string, results int) (string, error) {
	u, _ := url.Parse("https://api.search.brave.com/res/v1/news/search")
	q := u.Query()
	q.Set("q", query)
	q.Set("count", fmt.Sprintf("%d", results))
	//q.Set("country", "us")
	q.Set("search_lang", "en")
	q.Set("summary", "1")
	q.Set("extra_snippets", "1")
	//	q.Set("result_filter", "summarizer")
	u.RawQuery = q.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("X-Subscription-Token", b.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	braveResp := &BraveNewsResponse{}
	if err = json.Unmarshal(body, braveResp); err != nil {
		return "", err
	}

	// Convert to something simpler for the model to parse
	articles := make([]NewsResult, 0, len(braveResp.Results))
	for _, article := range braveResp.Results {
		articles = append(articles, NewsResult{
			Title:       article.Title,
			Description: article.Description,
			Summary:     strings.Join(article.ExtraSnippets, "\n"),
			//Link:        article.URL,
		})
	}

	data, err := json.Marshal(articles)
	return string(data), err
}

func NewBraveSearch(apiKey string) *BraveSearch {
	return &BraveSearch{
		apiKey: apiKey,
	}
}
