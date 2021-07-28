package assis

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/mail"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/gomarkdown/markdown"
	"github.com/gosimple/slug"
	"go.uber.org/zap"
)

type Tags []string

type Article struct {
	ID        string
	Permalink string
	Title     string
	Date      string
	Content   template.HTML
	Preview   template.HTML
	Template  string
	Pin       bool
	Published bool
	Tags      Tags
	Authors   []string
}

func newArticle(filename, relative string) (Article, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return Article{}, err
	}

	r := strings.NewReader(string(b))
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return Article{}, err
	}

	body, err := ioutil.ReadAll(msg.Body)
	if err != nil {
		return Article{}, err
	}

	id := slug.Make(msg.Header.Get("title"))

	pinned := false
	if msg.Header.Get("pin") == "true" {
		pinned = true
	}

	active := true
	if msg.Header.Get("active") == "false" {
		active = false
	}

	var tags []string
	if msg.Header.Get("tags") != "" {
		tags = strings.Split(msg.Header.Get("tags"), ",")
		for i := 0; i < len(tags); i++ {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	var authors []string
	if msg.Header.Get("authors") != "" {
		authors = strings.Split(msg.Header.Get("authors"), ",")
		for i := 0; i < len(authors); i++ {
			authors[i] = strings.TrimSpace(authors[i])
		}
	}

	return Article{
		ID:        id,
		Permalink: fmt.Sprintf("%s/%s.html", relative, id),
		Title:     msg.Header.Get("title"),
		Date:      msg.Header.Get("date"),
		Content:   template.HTML(markdown.ToHTML(body, nil, nil)),
		Preview:   template.HTML(markdown.ToHTML(body[:500], nil, nil)),
		Template:  msg.Header.Get("template"),
		Pin:       pinned,
		Published: active,
		Tags:      tags,
		Authors:   authors,
	}, nil
}

type ArticlePlugin struct {
	config    *Config
	templates map[string]*template.Template
	files     map[string][]Article
	name      string
	logger    *zap.Logger
}

func NewArticlePlugin(config *Config, logger *zap.Logger) *ArticlePlugin {
	return &ArticlePlugin{
		config:    config,
		templates: map[string]*template.Template{},
		files:     map[string][]Article{},
		name:      "markdown",
		logger:    logger,
	}
}

func (m ArticlePlugin) OnRegisterCustomFunction() map[string]interface{} {
	return map[string]interface{}{
		"articleCollection": m.articleCollection,
		"pinCollection":     m.pinCollection,
		"generateSearch":    m.generateSearch,
		"tags":              m.tags,
		"limit":             m.limit,
		"orderByDate":       m.orderByDate,
	}
}

func (m ArticlePlugin) generateSearch(filters []string) string {
	return strings.Join(filters, ",")
}

func (m ArticlePlugin) tags(articles []Article) Tags {
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

func (m ArticlePlugin) getCollection(path string, pin bool) []Article {
	for entry, collection := range m.files {
		if entry == m.config.Content+path {
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

func (m ArticlePlugin) articleCollection(path string) []Article {
	return m.getCollection(path, false)
}

func (m ArticlePlugin) pinCollection(path string) []Article {
	return m.getCollection(path, true)
}

func (m ArticlePlugin) limit(size int, list []Article) []Article {
	if len(list) == 0 {
		return []Article{}
	}
	if len(list) <= size {
		return list
	}
	return list[:size]
}

func (m ArticlePlugin) orderByDate(dir string, list []Article) []Article {
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

func (m ArticlePlugin) OnRender(t AssisTemplate, siteFiles SiteFiles, templates Templates) error {
	m.logger.Info("Start Article rendering")
	wp := workerpool.New(2)
	for _, container := range siteFiles {
		container := container
		wp.Submit(func() {
			if err := m.processContainer(container, t, templates); err != nil {
				m.logger.Error(err.Error())
			}
		})
	}
	wp.StopWait()
	m.logger.Info("Finished Article rendering")
	return nil
}

func (m ArticlePlugin) processContainer(container *FileContainer, t AssisTemplate, templates Templates) error {
	markdownFiles := container.FilterExt([]string{MD})
	for _, file := range markdownFiles {
		m.logger.Info("Read Article: " + container.FullFilename(file))
		rel, _ := filepath.Rel(m.config.Content, container.entry)
		parsed, err := newArticle(container.FullFilename(file), rel)
		if err != nil {
			return err
		}

		m.files[container.entry] = append(m.files[container.entry], parsed)

		output := strings.Replace(container.OutputFilename(file), string(file), parsed.ID+".html", 1)

		err = func() error {
			target, err := CreateTargetFile(output)
			defer target.Close()
			if err != nil {
				return err
			}

			templateFile, err := filepath.Abs(fmt.Sprintf("%s\\%s", m.config.Template.Path, parsed.Template))
			if err != nil {
				return err
			}

			targetTemplate, err := t.GetTemplate().ParseFiles(append(templates.baseOrdered, templateFile)...)
			if err != nil {
				return err
			}

			if err = targetTemplate.ExecuteTemplate(target, "layout", parsed); err != nil {
				return err
			}

			m.logger.Info("Rendered markdown to: " + target.Name())
			return nil
		}()

		if err != nil {
			return err
		}
	}
	return nil
}
