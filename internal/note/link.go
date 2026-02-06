package note

import (
	"regexp"
	"strings"
)

var linkPattern = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

// ExtractLinks はコンテンツから [[...]] リンクをすべて抽出する
func ExtractLinks(content string) []string {
	matches := linkPattern.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var links []string
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if name != "" && !seen[name] {
			seen[name] = true
			links = append(links, name)
		}
	}
	return links
}

// ResolveLinks はリンク名からノートを検索し、見つかったものと見つからなかったものを返す
func ResolveLinks(storage *Storage, links []string) (found []*Note, notFound []string) {
	for _, link := range links {
		n, err := storage.Find(link)
		if err != nil {
			notFound = append(notFound, link)
		} else {
			found = append(found, n)
		}
	}
	return
}

// FindBacklinks はtargetTitleを参照しているノートを検索する
func FindBacklinks(storage *Storage, targetTitle string) ([]*Note, error) {
	allNotes, err := storage.List("")
	if err != nil {
		return nil, err
	}

	var backlinks []*Note
	for _, n := range allNotes {
		links := ExtractLinks(n.Content)
		for _, link := range links {
			if strings.EqualFold(link, targetTitle) {
				backlinks = append(backlinks, n)
				break
			}
		}
	}
	return backlinks, nil
}
