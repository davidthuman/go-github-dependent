package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type DependentRepository struct {
	Owner  string
	Name   string
	After  string
	Before string
	Type   string // 'PACKAGE' or 'REPOSITORY'
	Stars  int
	Forks  int
}

type QueryDependentsConfig struct {
	MaxPages int
}

type HtmlPage struct {
	Page *html.Node
	Url  string
}

func containsKeyValue(elems []html.Attribute, key, value string) bool {
	for _, s := range elems {
		if key == s.Key && value == s.Val {
			return true
		}
	}
	return false
}

func parseDependentsPage(page HtmlPage) ([]DependentRepository, error) {

	slog.Debug("Parsing pages for dependent repos", "url", page.Url)

	urlParsed, _ := url.Parse(page.Url)
	params, _ := url.ParseQuery(urlParsed.RawQuery)

	after_list := params["dependents_after"]
	after := ""
	if len(after_list) != 0 {
		after = after_list[0]
	}
	before_list := params["dependents_before"]
	before := ""
	if len(before_list) != 0 {
		before = after_list[0]
	}
	type_list := params["dependent_type"]
	dep_type := "REPOSITORY"
	if len(type_list) != 0 {
		dep_type = after_list[0]
	}

	dependents := make([]DependentRepository, 0)

	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "div" && containsKeyValue(node.Attr, "data-test-id", "dg-repo-pkg-dependent") {

			//slog.Debug("At node:", node)

			// inside: image, span, div
			// inside span: a, a, small
			// first a: data-hovercard-type="organization"
			// second a: data-hovercard-type="repository"

			spanNode := node.FirstChild.NextSibling.NextSibling.NextSibling

			firstA := spanNode.FirstChild.NextSibling
			secondA := spanNode.FirstChild.NextSibling.NextSibling.NextSibling

			dependentRepo := DependentRepository{
				Owner:  firstA.FirstChild.Data,
				Name:   secondA.FirstChild.Data,
				After:  after,
				Before: before,
				Type:   dep_type,
			}

			// slog.Debug(dependentRepo)

			dependents = append(dependents, dependentRepo)
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(page.Page)

	return dependents, nil
}

func pageToDependents(pageCh <-chan HtmlPage, dependentsCh chan<- []DependentRepository, closeCh <-chan bool) {
	for {
		page, ok := <-pageCh
		if ok {
			dependents, err := parseDependentsPage(page)
			if err != nil {
				slog.Error("Error when parsing dependents page", "error", err)
				return
			}
			dependentsCh <- dependents
		} else {
			slog.Debug("Page Channel closed, closing Dependents Channel")
			// close(dependentsCh)
			return
		}

		//case <-closeCh:
		//	slog.Debug("Received close witin pageToDependents")
		//	return

	}

}

func parseDependentsPageForNextUrl(doc *html.Node) (string, error) {

	slog.Debug("Parsing page for next URL")

	nextPageURL := ""

	var crawler func(*html.Node)
	crawler = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "div" && containsKeyValue(node.Attr, "class", "paginate-container") {
			for _, attr := range node.FirstChild.NextSibling.FirstChild.NextSibling.Attr {
				if attr.Key == "href" {
					nextPageURL = attr.Val
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			crawler(child)
		}
	}
	crawler(doc)

	slog.Debug("Next Page URL", "next_page_url", nextPageURL)

	return nextPageURL, nil

}

func parsePageToNextUrl(pageCh <-chan *html.Node, nextUrlCh chan<- string, closeCh chan<- bool, maxPages int) {

	pageCount := 0

	for {
		if pageCount >= maxPages {
			break
		}
		doc := <-pageCh
		nextUrl, err := parseDependentsPageForNextUrl(doc)
		if err != nil {
			slog.Debug("Error when paring page for next url", "error", err)
			return
		}
		pageCount += 1
		slog.Debug("Current page count", "count", pageCount)
		nextUrlCh <- nextUrl

	}

	slog.Debug("Closing channels")
	close(nextUrlCh)
	//closeCh <- true

}

func getDependentsPage(url string) (*html.Node, error) {

	slog.Debug("Requesting GitHub dependent page", "url", url)

	response, err := http.Get(url)

	if err != nil {
		slog.Error("Error when making request to GitHub", "error", err, "url", url)
		return nil, err
	}

	defer response.Body.Close()

	slog.Debug("Response", "status_code", response.StatusCode)

	if response.StatusCode != 200 {
		slog.Error("Response status code is not 200", "status_code", response.StatusCode, "status", response.Status)
		log.Println(response.Header)
		body, _ := io.ReadAll(response.Body)
		log.Println(string(body))
		return nil, errors.New("response status code is not 200")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("Error when reading response body", "error", err)
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(string(body)))

	if err != nil {
		slog.Error("Error when parsing HTML", "error", err)
		return nil, err
	}

	return doc, nil
}

func urlToPage(urlCh <-chan string, pageCh chan<- HtmlPage, closeCh <-chan bool) {

	for {
		nextUrl, ok := <-urlCh
		if ok {
			slog.Debug("Received URL in urlToPage", "url", nextUrl)
			doc, err := getDependentsPage(nextUrl)
			if err != nil {
				slog.Error("Error when getting dependents page", "error", err, "url", nextUrl)
				return
			}
			pageCh <- HtmlPage{Page: doc, Url: nextUrl}
		} else {
			slog.Debug("URL Channel closed, closing page channel")
			close(pageCh)
			return
		}
		//case <-closeCh:
		//	slog.Debug("Received close within urlToPage")
		//	return

	}

}

func GetDependents(repoOwner, repoName string, config QueryDependentsConfig) ([]DependentRepository, error) {

	// url: https://github.com/inconshreveable/mousetrap/network/dependents
	// pagination: ?dependents_after=NDgwMjE1MDcyNjI

	// ?
	// owner=golang
	// dependents_after=...
	// dependents_before=...
	// dependent_type=REPOSITORY
	// dependent_type=PACKAGE

	// Each page shows 30 repos

	// NDgwMjI2ODkxMjU - 48022689125
	// NDgwMTg2Mjg4MTA - 48018628810

	// NDgwMjE1MDcyNjI - 48021507262
	// NDgwMTg1OTU2NDM - 48018595643

	// repo: ajaypvictor/cloud-api-adaptor
	// NDgwMTg1OTU2NDM - 48018595643
	// curl -H "Authorization: bearer TOKEN" \
	// -X POST \
	// -d '{"query": "query { repository(owner: \"ajaypvictor\", name: \"cloud-api-adaptor\") { id databaseId } }"}' \
	// https://api.github.com/graphql

	// {"data":{"repository":{"id":"R_kgDOLmh-qA","databaseId":778600104}}}

	// div
	// data-test-id="dg-repo-pkg-dependent"

	pagesProcessed := 0

	nextPageURL := fmt.Sprintf("https://github.com/%s/%s/network/dependents", repoOwner, repoName)

	allDependents := make([]DependentRepository, 0)

	for nextPageURL != "" && pagesProcessed < config.MaxPages {

		doc, err := getDependentsPage(nextPageURL)

		if err != nil {
			slog.Error("Error when getting dependents page", "error", err)
			return nil, err
		}

		dependents, err := parseDependentsPage(HtmlPage{Page: doc, Url: nextPageURL})

		if err != nil {
			slog.Error("Error when parsing dependents page", "error", err)
			return nil, err
		}

		nextPageURL, _ = parseDependentsPageForNextUrl(doc)

		allDependents = append(allDependents, dependents...)

		pagesProcessed += 1

	}

	// slog.Debug(allDependents)

	return allDependents, nil

}

func GetDependentsProducerConsumer(repoOwner, repoName string, config QueryDependentsConfig) ([]DependentRepository, error) {

	// URL to Page
	// Page to Next URL

	// Page to list of Dependents

	urlCh := make(chan string, config.MaxPages)
	pageCh := make(chan HtmlPage, config.MaxPages)
	pageForUrlCh := make(chan *html.Node, config.MaxPages)
	pageForDependentsCh := make(chan HtmlPage, config.MaxPages)
	dependentsCh := make(chan []DependentRepository, config.MaxPages)
	closeCh := make(chan bool)

	//var wg sync.WaitGroup

	urlCh <- fmt.Sprintf("https://github.com/%s/%s/network/dependents", repoOwner, repoName)

	go urlToPage(urlCh, pageCh, closeCh)

	go parsePageToNextUrl(pageForUrlCh, urlCh, closeCh, config.MaxPages)

	go pageToDependents(pageForDependentsCh, dependentsCh, closeCh)

	dependents := make([]DependentRepository, 0)

out:
	for {
		select {
		case htmlPage, ok := <-pageCh:
			if ok {
				pageForUrlCh <- htmlPage.Page
				pageForDependentsCh <- htmlPage
			} else {
				slog.Debug("Page Channel closed")
				close(pageForUrlCh)
				close(pageForDependentsCh)
				break out
			}
		case newDependents := <-dependentsCh:
			slog.Debug("Received dependents list", "length", len(newDependents))
			dependents = append(dependents, newDependents...)
			//case <-closeCh:
			//	break out
		}
	}

	slog.Debug("OUT of for loop")

	return dependents, nil

}

func main() {
	//getRepo()

	slog.SetLogLoggerLevel(slog.LevelDebug)

	GetDependentsProducerConsumer("inconshreveable", "mousetrap", QueryDependentsConfig{MaxPages: 5})
}
