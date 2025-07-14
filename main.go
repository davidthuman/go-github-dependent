package main

import (
	"fmt"
	"io"
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

func containsKeyValue(elems []html.Attribute, key, value string) bool {
	for _, s := range elems {
		if key == s.Key && value == s.Val {
			return true
		}
	}
	return false
}

func parseDependentPage(doc *html.Node, urlFrom string) ([]DependentRepository, string, error) {

	urlParsed, _ := url.Parse(urlFrom)
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

	nextPageURL := ""

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

	return dependents, nextPageURL, nil

}

func getDependentsPage(url string) (*html.Node, error) {

	response, err := http.Get(url)

	if err != nil {
		slog.Error("Error when making request to GitHub", "error", err)
		return nil, err
	}

	defer response.Body.Close()

	slog.Debug("Response", "status_code", response.StatusCode)

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

		dependents, newNextPageURL, err := parseDependentPage(doc, nextPageURL)

		if err != nil {
			slog.Error("Error when parsing dependents page", "error", err)
			return nil, err
		}

		nextPageURL = newNextPageURL

		allDependents = append(allDependents, dependents...)

		pagesProcessed += 1

	}

	// slog.Debug(allDependents)

	return allDependents, nil

}

func main() {
	//getRepo()

	slog.SetLogLoggerLevel(slog.LevelInfo)

	GetDependents("inconshreveable", "mousetrap", QueryDependentsConfig{MaxPages: 5})
}
