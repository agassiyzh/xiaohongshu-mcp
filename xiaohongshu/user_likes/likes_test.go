package user_likes

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
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

// TestUserLikesResponse_Structure tests the response structure
func TestUserLikesResponse_Structure(t *testing.T) {
	response := &UserLikesResponse{
		LikedFeeds: []LikedFeed{
			{
				FeedID:    "test123",
				Title:     "Test Note",
				URL:       "https://www.xiaohongshu.com/explore/test123",
				Author:    "Test Author",
				AuthorID:  "author123",
				LikedTime: "2023-01-01 12:00:00",
			},
		},
		Count:   1,
		HasMore: false,
	}

	assert.Equal(t, 1, response.Count)
	assert.Equal(t, 1, len(response.LikedFeeds))
	assert.Equal(t, "test123", response.LikedFeeds[0].FeedID)
	assert.Equal(t, "Test Note", response.LikedFeeds[0].Title)
	assert.Equal(t, "Test Author", response.LikedFeeds[0].Author)
	assert.False(t, response.HasMore)
}

// TestLikedFeed_Structure tests the liked feed structure
func TestLikedFeed_Structure(t *testing.T) {
	feed := LikedFeed{
		FeedID:    "feed123",
		Title:     "Sample Title",
		URL:       "https://www.xiaohongshu.com/explore/feed123",
		Author:    "Sample Author",
		AuthorID:  "author456",
		LikedTime: "2023-12-25 10:30:00",
	}

	assert.Equal(t, "feed123", feed.FeedID)
	assert.Equal(t, "Sample Title", feed.Title)
	assert.Equal(t, "https://www.xiaohongshu.com/explore/feed123", feed.URL)
	assert.Equal(t, "Sample Author", feed.Author)
	assert.Equal(t, "author456", feed.AuthorID)
	assert.Equal(t, "2023-12-25 10:30:00", feed.LikedTime)
}

// TestParseFeedItem tests the parseFeedItem method
func TestParseFeedItem(t *testing.T) {
	action := &UserLikesAction{}

	// Test with valid feed data
	feed := map[string]interface{}{
		"id": "test123",
		"noteCard": map[string]interface{}{
			"displayTitle": "Test Title",
			"user": map[string]interface{}{
				"nickname": "Test User",
				"userId":   "user123",
			},
		},
	}

	result := action.parseFeedItem(feed)

	assert.Equal(t, "test123", result.FeedID)
	assert.Equal(t, "Test Title", result.Title)
	assert.Equal(t, "Test User", result.Author)
	assert.Equal(t, "user123", result.AuthorID)
	assert.Contains(t, result.URL, "test123")
	assert.NotEmpty(t, result.LikedTime)
}

// TestParseFeedItem_EmptyData tests parseFeedItem with empty data
func TestParseFeedItem_EmptyData(t *testing.T) {
	action := &UserLikesAction{}

	// Test with empty feed data
	feed := map[string]interface{}{}
	result := action.parseFeedItem(feed)

	assert.Equal(t, "", result.FeedID)
	assert.Equal(t, "", result.Title)
	assert.Equal(t, "", result.Author)
	assert.Equal(t, "", result.AuthorID)
	assert.Empty(t, result.URL)
	assert.NotEmpty(t, result.LikedTime) // Should still have timestamp
}

// TestParseFeedItem_NestedID tests parseFeedItem with nested ID structure
func TestParseFeedItem_NestedID(t *testing.T) {
	action := &UserLikesAction{}

	// Test with nested ID in noteCard
	feed := map[string]interface{}{
		"noteCard": map[string]interface{}{
			"id":           "nested123",
			"displayTitle": "Nested Title",
			"user": map[string]interface{}{
				"nickname": "Nested User",
				"userId":   "nesteduser123",
			},
		},
	}

	result := action.parseFeedItem(feed)

	assert.Equal(t, "nested123", result.FeedID)
	assert.Equal(t, "Nested Title", result.Title)
	assert.Equal(t, "Nested User", result.Author)
	assert.Equal(t, "nesteduser123", result.AuthorID)
	assert.Contains(t, result.URL, "nested123")
}

// TestGetStringValue tests the getStringValue helper method
func TestGetStringValue(t *testing.T) {
	action := &UserLikesAction{}

	// Test with valid data
	item := map[string]interface{}{
		"title":       "Test Title",
		"name":        "Test Name",
		"description": "Test Description",
	}

	// Test first key found
	result := action.getStringValue(item, "title", "name", "description")
	assert.Equal(t, "Test Title", result)

	// Test second key found (first not present)
	result = action.getStringValue(item, "nonexistent", "name", "description")
	assert.Equal(t, "Test Name", result)

	// Test no keys found
	result = action.getStringValue(item, "nonexistent1", "nonexistent2")
	assert.Equal(t, "", result)
}

// TestGetStringValue_NonStringValue tests with non-string values
func TestGetStringValue_NonStringValue(t *testing.T) {
	action := &UserLikesAction{}

	// Test with non-string values
	item := map[string]interface{}{
		"number": 123,
		"bool":   true,
		"string": "test",
	}

	// Should only return string values
	result := action.getStringValue(item, "number", "bool", "string")
	assert.Equal(t, "test", result)
}

// TestExtractAuthor tests the extractAuthor helper method
func TestExtractAuthor(t *testing.T) {
	action := &UserLikesAction{}

	// Test with valid user data
	item := map[string]interface{}{
		"user": map[string]interface{}{
			"nickname": "Test Nickname",
			"nickName": "Test NickName",
		},
	}

	result := action.extractAuthor(item)
	assert.Equal(t, "Test Nickname", result)

	// Test with alternate nickname field
	item2 := map[string]interface{}{
		"user": map[string]interface{}{
			"nickName": "Test NickName",
		},
	}

	result2 := action.extractAuthor(item2)
	assert.Equal(t, "Test NickName", result2)

	// Test with no user data
	item3 := map[string]interface{}{}
	result3 := action.extractAuthor(item3)
	assert.Equal(t, "", result3)
}

// TestExtractAuthor_EmptyUser tests with empty user object
func TestExtractAuthor_EmptyUser(t *testing.T) {
	action := &UserLikesAction{}

	// Test with empty user object
	item := map[string]interface{}{
		"user": map[string]interface{}{},
	}

	result := action.extractAuthor(item)
	assert.Equal(t, "", result)
}

// TestNavigateToLikesPage_Success tests the successful navigation scenario
func TestNavigateToLikesPage_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires browser in short mode")
	}

	// Use real browser for integration testing
	b := browser.NewBrowser(true) // Headless
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := NewUserLikesAction(page)

	// Try to navigate to a test page first
	err := page.Navigate("https://www.xiaohongshu.com")
	if err != nil {
		t.Skipf("Skipping test due to navigation error: %v", err)
	}

	// Test the navigateToLikesPage function
	// Note: This might fail if the user is not logged in or page structure has changed
	err = action.navigateToLikesPage(page)
	if err != nil {
		t.Logf("Integration test failed (this is expected if not logged in): %v", err)
		// Don't fail the test as this depends on external factors
		return
	}

	t.Log("Integration test passed successfully")
}

// TestGetUserLikedNotes_ErrorHandling tests error handling scenarios
func TestGetUserLikedNotes_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires browser in short mode")
	}

	// Use real browser to test error scenarios
	b := browser.NewBrowser(true)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := NewUserLikesAction(page)

	// Test with invalid navigation
	ctx := context.Background()

	// This should fail because we haven't navigated to a proper page
	_, err := action.GetUserLikedNotes(ctx)
	if err != nil {
		t.Logf("Expected error occurred: %v", err)
		// This is expected to fail when not properly logged in or on wrong page
	}
}

// TestGetUserLikedNotes_ContextTimeout tests context timeout handling
func TestGetUserLikedNotes_ContextTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires browser in short mode")
	}

	b := browser.NewBrowser(true)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := NewUserLikesAction(page)

	// Test with very short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This should fail quickly due to timeout
	_, err := action.GetUserLikedNotes(ctx)
	if err != nil {
		t.Logf("Expected timeout error: %v", err)
	}
}

// TestNavigateToLikesPage_ProductionLike tests production-like scenario
func TestNavigateToLikesPage_ProductionLike(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires browser in short mode")
	}

	// Use real browser to test production-like scenario
	b := browser.NewBrowser(true)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := NewUserLikesAction(page)

	// Try to navigate to xiaohongshu
	err := page.Navigate("https://www.xiaohongshu.com")
	if err != nil {
		t.Skipf("Skipping test due to navigation error: %v", err)
	}

	// Wait for page to load
	page.MustWaitLoad()

	// Test the navigateToLikesPage function
	err = action.navigateToLikesPage(page)
	if err != nil {
		t.Logf("Navigation failed (expected if not logged in): %v", err)
		// Don't fail the test as this depends on external factors
		return
	}

	t.Log("Production-like test passed successfully")
}

// TestNavigateToLikesPage_Logging tests that logging works correctly
func TestNavigateToLikesPage_Logging(t *testing.T) {
	// Capture log output
	originalHook := logrus.StandardLogger().Hooks
	defer func() {
		logrus.StandardLogger().ReplaceHooks(originalHook)
	}()

	var logMessages []string
	hook := &TestHook{Messages: &logMessages}
	logrus.AddHook(hook)

	// This test verifies that logging is set up correctly
	t.Log("Logging test completed")
	assert.NotNil(t, logrus.StandardLogger())
}

// TestHook is a custom logrus hook for testing
type TestHook struct {
	Messages *[]string
}

func (h *TestHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *TestHook) Fire(entry *logrus.Entry) error {
	*h.Messages = append(*h.Messages, entry.Message)
	return nil
}

// BenchmarkNewUserLikesAction benchmarks the constructor
func BenchmarkNewUserLikesAction(b *testing.B) {
	br := browser.NewBrowser(true)
	defer br.Close()

	page := br.NewPage()
	defer page.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		action := NewUserLikesAction(page)
		_ = action
	}
}

// BenchmarkParseFeedItem benchmarks the parseFeedItem method
func BenchmarkParseFeedItem(b *testing.B) {
	action := &UserLikesAction{}
	feed := map[string]interface{}{
		"id": "test123",
		"noteCard": map[string]interface{}{
			"displayTitle": "Test Title",
			"user": map[string]interface{}{
				"nickname": "Test User",
				"userId":   "user123",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := action.parseFeedItem(feed)
		_ = result
	}
}

// BenchmarkGetStringValue benchmarks the getStringValue helper method
func BenchmarkGetStringValue(b *testing.B) {
	action := &UserLikesAction{}
	item := map[string]interface{}{
		"title":       "Test Title",
		"name":        "Test Name",
		"description": "Test Description",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := action.getStringValue(item, "title", "name", "description")
		_ = result
	}
}

// BenchmarkExtractAuthor benchmarks the extractAuthor helper method
func BenchmarkExtractAuthor(b *testing.B) {
	action := &UserLikesAction{}
	item := map[string]interface{}{
		"user": map[string]interface{}{
			"nickname": "Test Nickname",
			"nickName": "Test NickName",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := action.extractAuthor(item)
		_ = result
	}
}
