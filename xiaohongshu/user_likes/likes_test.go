package user_likes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xpzouying/xiaohongshu-mcp/browser"
)

// TestNewUserLikesAction tests the constructor
func TestNewUserLikesAction(t *testing.T) {
	// Test with nil page
	action := NewUserLikesAction(nil)
	assert.NotNil(t, action)
	assert.Nil(t, action.page)

	// Test with valid page
	b := browser.NewBrowser(true)
	defer b.Close()
	page := b.NewPage()
	defer page.Close()

	action = NewUserLikesAction(page)

	assert.NotNil(t, action)
	assert.NotNil(t, action.page)
}

func TestGotoProfilePage(t *testing.T) {

	b := browser.NewBrowser(false)
	defer b.Close()
	page := b.NewPage()
	defer page.Close()

	action := NewUserLikesAction(page)

	action.navigateToLikesPage(page)
}
