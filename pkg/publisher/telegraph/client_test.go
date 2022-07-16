package telegraph

import (
	"encoding/json"
	"testing"

	"arhat.dev/mbot/pkg/rt"
	"github.com/stretchr/testify/assert"
)

func TestNode(t *testing.T) {
	const (
		TEST_DATA = `["test-text",{"tag":"p","children":["test-text"]}]`
	)

	expected := []telegraphNode{
		{
			Text: "test-text",
		},
		{
			Elm: rt.NewOptionalValue(telegraphNodeElement{
				Tag: "p",
				Children: []telegraphNode{
					{
						Text: "test-text",
					},
				},
			}),
		},
	}
	data, err := json.Marshal(expected)
	assert.NoError(t, err)
	assert.EqualValues(t, []byte(TEST_DATA), data)

	var actual []telegraphNode
	err = json.Unmarshal([]byte(TEST_DATA), &actual)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
}

func TestClient(t *testing.T) {
	const (
		TEST_SHORT_NAME  = "test-short-name"
		TEST_AUTHOR_NAME = "test-author-name"
		TEST_AUTHOR_URL  = "https://example.com/test-author-url"
		TEST_PAGE_TITLE  = "test-page-title"
	)
	c, err := newDefaultClient()
	if !assert.NoError(t, err) {
		return
	}

	account, err := c.CreateAccount(createAccountOptions{
		ShortName:  TEST_SHORT_NAME,
		AuthorName: TEST_AUTHOR_NAME,
		AuthorURL:  TEST_AUTHOR_URL,
	})
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, len(account.AccessToken), 0)
	assert.Greater(t, len(account.AuthURL), 0)
	assert.Equal(t, telegraphAccount{
		ShortName:   TEST_SHORT_NAME,
		AuthorName:  TEST_AUTHOR_NAME,
		AuthorURL:   TEST_AUTHOR_URL,
		AccessToken: account.AccessToken,
		AuthURL:     account.AuthURL,
		PageCount:   0,
	}, account)

	var TEST_CONTENT = []telegraphNode{
		{
			Elm: rt.NewOptionalValue(telegraphNodeElement{
				Tag: "p",
				Children: []telegraphNode{
					{
						Text: "test-content",
					},
				},
			}),
		},
		{
			Text: "foo",
		},
	}

	page, err := c.CreatePage(createPageOptions{
		AccessToken:   account.AccessToken,
		Title:         TEST_PAGE_TITLE,
		AuthorName:    TEST_AUTHOR_NAME,
		AuthorURL:     TEST_AUTHOR_URL,
		Content:       TEST_CONTENT,
		ReturnContent: true,
	})
	if !assert.NoError(t, err) {
		return
	}
	assert.Greater(t, len(page.Path), 0)
	assert.Greater(t, len(page.URL), 0)
	page.Views = 0
	assert.Equal(t, telegraphPage{
		Path:        page.Path,
		URL:         page.URL,
		Title:       TEST_PAGE_TITLE,
		Description: "",
		AuthorName:  TEST_AUTHOR_NAME,
		AuthorURL:   TEST_AUTHOR_URL,
		ImageURL:    "",
		Content:     TEST_CONTENT,
		Views:       0,
		CanEdit:     true,
	}, page)

	list, err := c.GetPageList(getPageListOptions{
		AccessToken: account.AccessToken,
		Offset:      0,
		Limit:       10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), list.TotalCount)
	if assert.Equal(t, 1, len(list.Pages)) {
		list.Pages[0].Description = "" // strip server generated description
		assert.Equal(t, telegraphPage{
			Path:        page.Path,
			URL:         page.URL,
			Title:       TEST_PAGE_TITLE,
			Description: "",
			AuthorName:  TEST_AUTHOR_NAME,
			AuthorURL:   TEST_AUTHOR_URL,
			ImageURL:    "",
			Content:     nil,
			Views:       0,
			CanEdit:     true,
		}, list.Pages[0])
	}

	const (
		EDITED_PAGE_TITLE  = "edited-title"
		EDITED_AUTHOR_NAME = "edited-author-name"
		EDITED_AUTHOR_URL  = "https://example.com/edited-author-url"
	)
	var EDITED_CONTENT = []telegraphNode{
		{
			Elm: rt.NewOptionalValue(telegraphNodeElement{
				Tag: "h3",
				Attrs: map[string]string{
					// telegraph server will set this attr
					"id": "test-content-edited",
				},
				Children: []telegraphNode{
					{
						Text: "test-content-edited",
					},
				},
			}),
		},
		{
			Text: "bar",
		},
	}
	editedPage, err := c.EditPage(page.Path, createPageOptions{
		AccessToken:   account.AccessToken,
		Title:         EDITED_PAGE_TITLE,
		AuthorName:    EDITED_AUTHOR_NAME,
		AuthorURL:     EDITED_AUTHOR_URL,
		Content:       EDITED_CONTENT,
		ReturnContent: true,
	})
	assert.NoError(t, err)
	editedPage.Views = 0
	assert.Equal(t, telegraphPage{
		Path:        page.Path,
		URL:         page.URL,
		Title:       EDITED_PAGE_TITLE,
		Description: "",
		AuthorName:  EDITED_AUTHOR_NAME,
		AuthorURL:   EDITED_AUTHOR_URL,
		ImageURL:    "",
		Content:     EDITED_CONTENT,
		Views:       0,
		CanEdit:     true,
	}, editedPage)
}
