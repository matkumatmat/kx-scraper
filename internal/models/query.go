package models

import (
	"fmt"
	"strings"
)

type PostsSearchParams struct {
	MainPhrase    string
	ExactPhrase   string
	AnyPhrase     string
	ExcludePhrase string
	HashtagPhrase string
	FromAccount   string
	ToAccount     string
	Mentioning    string
	MinReplies    int
	MinFaves      int
	MinRetweets   int
	Lang          string
	Until         string
	Since         string
}

func (p *PostsSearchParams) BuildQuery() string {
	var parts []string
	if p.MainPhrase != "" {
		parts = append(parts, p.MainPhrase)
	}
	if p.ExactPhrase != "" {
		parts = append(parts, fmt.Sprintf(`"%s"`, p.ExactPhrase))
	}
	if p.AnyPhrase != "" {
		anyParts := strings.Split(p.AnyPhrase, " ")
		for _, ap := range anyParts {
			parts = append(parts, fmt.Sprintf(`(%s)`, ap))
		}
	}
	if p.ExcludePhrase != "" {
		excludeParts := strings.Split(p.ExcludePhrase, " ")
		for _, ep := range excludeParts {
			parts = append(parts, fmt.Sprintf(`-%s`, ep))
		}
	}
	if p.HashtagPhrase != "" {
		hashtagParts := strings.Split(p.HashtagPhrase, " ")
		for _, hp := range hashtagParts {
			parts = append(parts, fmt.Sprintf(`#%s`, hp))
		}
	}
	if p.FromAccount != "" {
		parts = append(parts, fmt.Sprintf(`from:%s`, p.FromAccount))
	}
	if p.ToAccount != "" {
		parts = append(parts, fmt.Sprintf(`to:%s`, p.ToAccount))
	}
	if p.Mentioning != "" {
		parts = append(parts, fmt.Sprintf(`@%s`, p.Mentioning))
	}
	if p.MinReplies > 0 {
		parts = append(parts, fmt.Sprintf(`min_replies:%d`, p.MinReplies))
	}
	if p.MinFaves > 0 {
		parts = append(parts, fmt.Sprintf(`min_faves:%d`, p.MinFaves))
	}
	if p.MinRetweets > 0 {
		parts = append(parts, fmt.Sprintf(`min_retweets:%d`, p.MinRetweets))
	}
	if p.Lang != "" {
		parts = append(parts, fmt.Sprintf(`lang:%s`, p.Lang))
	}
	if p.Until != "" {
		parts = append(parts, fmt.Sprintf(`until:%s`, p.Until))
	}
	if p.Since != "" {
		parts = append(parts, fmt.Sprintf(`since:%s`, p.Since))
	}
	return strings.Join(parts, " ")
}