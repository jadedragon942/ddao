package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jadedragon942/ddao/orm"
)

type WikiService struct {
	orm *orm.ORM
}

func NewWikiService(orm *orm.ORM) *WikiService {
	return &WikiService{orm: orm}
}

func (w *WikiService) CreatePage(title, content, authorID string) (*WikiPage, error) {
	pageID, err := generateID()
	if err != nil {
		return nil, err
	}

	page := &WikiPage{
		ID:        pageID,
		Title:     title,
		Content:   content,
		AuthorID:  authorID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	obj := wikiPageToObject(page)
	_, _, err = w.orm.Insert(context.Background(), obj)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (w *WikiService) GetPage(pageID string) (*WikiPage, error) {
	obj, err := w.orm.FindByID(context.Background(), "wiki_pages", pageID)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.New("page not found")
	}

	return objectToWikiPage(obj), nil
}

func (w *WikiService) GetPageByTitle(title string) (*WikiPage, error) {
	obj, err := w.orm.FindByKey(context.Background(), "wiki_pages", "title", title)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.New("page not found")
	}

	return objectToWikiPage(obj), nil
}

func (w *WikiService) UpdatePage(pageID, title, content string) (*WikiPage, error) {
	page, err := w.GetPage(pageID)
	if err != nil {
		return nil, err
	}

	page.Title = title
	page.Content = content
	page.UpdatedAt = time.Now()

	obj := wikiPageToObject(page)
	_, err = w.orm.Storage.Update(context.Background(), obj)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (w *WikiService) DeletePage(pageID string) error {
	_, err := w.orm.DeleteByID(context.Background(), "wiki_pages", pageID)
	return err
}

func (w *WikiService) SearchPages(query string) ([]*WikiPage, error) {
	if strings.TrimSpace(query) == "" {
		return []*WikiPage{}, nil
	}

	results := []*WikiPage{}

	searchTerms := strings.Fields(strings.ToLower(query))
	if len(searchTerms) == 0 {
		return results, nil
	}

	return results, nil
}

func (w *WikiService) GetAllPages() ([]*WikiPage, error) {
	return []*WikiPage{}, nil
}

func (w *WikiService) GetPagesByAuthor(authorID string) ([]*WikiPage, error) {
	results := []*WikiPage{}
	return results, nil
}