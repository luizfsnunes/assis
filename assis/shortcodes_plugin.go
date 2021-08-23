package assis

import (
	"html/template"
	"sort"
	"strings"
	"time"
)

type ShortcodesPlugin struct {
	Articles map[string][]Article
	config   *Config
}

func (sp ShortcodesPlugin) OnRegisterCustomFunction() map[string]interface{} {
	return map[string]interface{}{
		"truncate":          sp.Truncate,
		"articleCollection": sp.ArticleCollection,
		"pinCollection":     sp.PinCollection,
		"generateSearch":    sp.GenerateSearch,
		"tags":              sp.Tags,
		"limit":             sp.Limit,
		"orderByDate":       sp.OrderByDate,
	}
}

func (ShortcodesPlugin) Truncate(size int, str template.HTML) template.HTML {
	s := string(str)
	var numRunes = 0
	for index, _ := range s {
		numRunes++
		if numRunes > size {
			return template.HTML(s[:index] + "...")
		}
	}
	return str
}

func (ShortcodesPlugin) GenerateSearch(filters []string) string {
	return strings.Join(filters, ",")
}

func (ShortcodesPlugin) Tags(articles []Article) Tags {
	allTags := map[string]string{}
	for _, article := range articles {
		for _, tag := range article.Tags {
			if _, ok := allTags[tag]; !ok {
				allTags[tag] = tag
			}
		}
	}

	tags := Tags{}
	for _, tag := range allTags {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func (sp ShortcodesPlugin) GetCollection(path string, pin bool) []Article {
	for entry, collection := range sp.Articles {
		if entry == sp.config.Content+path {
			var out []Article
			for _, f := range collection {
				if f.Pin == pin && f.Published {
					out = append(out, f)
				}
			}
			return out
		}
	}
	return []Article{}
}

func (sp ShortcodesPlugin) ArticleCollection(path string) []Article {
	return sp.GetCollection(path, false)
}

func (sp ShortcodesPlugin) PinCollection(path string) []Article {
	return sp.GetCollection(path, true)
}

func (ShortcodesPlugin) Limit(size int, list []Article) []Article {
	if len(list) == 0 {
		return []Article{}
	}
	if len(list) <= size {
		return list
	}
	return list[:size]
}

func (ShortcodesPlugin) OrderByDate(dir string, list []Article) []Article {
	tmpList := list
	sort.Slice(tmpList, func(i, j int) bool {
		d1, _ := time.Parse("2006-01-02", list[i].Date)
		d2, _ := time.Parse("2006-01-02", list[j].Date)

		if dir == "desc" {
			return d1.Before(d2)
		}
		return d1.After(d2)
	})
	return tmpList
}
