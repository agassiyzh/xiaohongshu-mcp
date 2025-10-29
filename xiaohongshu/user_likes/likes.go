package user_likes

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/xiaohongshu"
)

// LikedFeed 表示用户点赞的笔记信息
type LikedFeed struct {
	FeedID    string `json:"feed_id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Author    string `json:"author"`
	AuthorID  string `json:"author_id"`
	LikedTime string `json:"liked_time"`
}

// UserLikesResponse 用户点赞笔记的响应
type UserLikesResponse struct {
	LikedFeeds []LikedFeed `json:"liked_feeds"`
	Count      int         `json:"count"`
	HasMore    bool        `json:"has_more"`
}

// UserLikesAction 获取用户点赞的笔记
type UserLikesAction struct {
	page *rod.Page
}

// NewUserLikesAction 创建 UserLikesAction 实例
func NewUserLikesAction(page *rod.Page) *UserLikesAction {
	action := &UserLikesAction{
		page: page,
	}
	// 只有当 page 不为 nil 时才设置 timeout
	if page != nil {
		action.page = page.Timeout(60 * time.Second)
	}
	return action
}

// GetUserLikedNotes 获取用户点赞的所有笔记
func (u *UserLikesAction) GetUserLikedNotes(ctx context.Context) (*UserLikesResponse, error) {
	page := u.page.Context(ctx)

	// 1. 首先导航到个人主页的点赞标签页
	if err := u.navigateToLikesPage(page); err != nil {
		return nil, fmt.Errorf("failed to navigate to likes page: %w", err)
	}

	// 2. 等待页面加载完成
	page.MustWaitStable()
	time.Sleep(2 * time.Second)

	// 3. 获取点赞的笔记数据
	likedNotes, err := u.extractLikedNotes(page)
	if err != nil {
		return nil, fmt.Errorf("failed to extract liked notes: %w", err)
	}

	return likedNotes, nil
}

// navigateToLikesPage 导航到用户个人主页的点赞标签页
func (u *UserLikesAction) navigateToLikesPage(page *rod.Page) error {
	// 首先访问用户个人主页

	navigation := xiaohongshu.NewNavigate(page)
	navigation.ToUserLikesPage(page.GetContext())

	return nil
}

// extractLikedNotes 从页面中提取点赞的笔记数据
func (u *UserLikesAction) extractLikedNotes(page *rod.Page) (*UserLikesResponse, error) {
	// 等待页面数据加载
	page.MustWait(`() => {
		// 等待 __INITIAL_STATE__ 或者页面内容加载
		return window.__INITIAL_STATE__ !== undefined ||
		       document.querySelector('.feeds-container') !== null ||
		       document.querySelector('.note-item') !== null;
	}`)

	// 尝试从 __INITIAL_STATE__ 获取数据
	if likedFeeds, err := u.extractFromInitialState(page); err == nil {
		return likedFeeds, nil
	}

	logrus.Warn("Failed to extract from __INITIAL_STATE__, trying DOM parsing")

	// 如果 __INITIAL_STATE__ 不可用，则从 DOM 解析
	return u.extractFromDOM(page)
}

// extractFromInitialState 从 __INITIAL_STATE__ 提取数据
func (u *UserLikesAction) extractFromInitialState(page *rod.Page) (*UserLikesResponse, error) {
	result := page.MustEval(`() => {
		try {
			if (window.__INITIAL_STATE__) {
				// 尝试多个可能的数据路径
				const possiblePaths = [
					'user.likedNotes.value',
					'user.likedNotes._value',
					'user.likes.value',
					'user.likes._value',
					'note.likedNotes.value',
					'note.likedNotes._value'
				];

				for (const path of possiblePaths) {
					const parts = path.split('.');
					let data = window.__INITIAL_STATE__;

					for (const part of parts) {
						if (data && data[part] !== undefined) {
							data = data[part];
						} else {
							data = null;
							break;
						}
					}

					if (data && (data.value !== undefined ? data.value : data._value)) {
						const notes = data.value !== undefined ? data.value : data._value;
						if (Array.isArray(notes) && notes.length > 0) {
							return JSON.stringify({
								data: notes,
								source: path
							});
						}
					}
				}
			}
			return null;
		} catch (error) {
			console.error('Error extracting from __INITIAL_STATE__:', error);
			return null;
		}
	}`).String()

	if result == "" || result == "null" {
		return nil, fmt.Errorf("no liked notes data found in __INITIAL_STATE__")
	}

	var response struct {
		Data []json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initial state data: %w", err)
	}

	// 解析每个笔记项
	var likedFeeds []LikedFeed
	for _, item := range response.Data {
		var feed map[string]interface{}
		if err := json.Unmarshal(item, &feed); err != nil {
			continue
		}

		likedFeed := u.parseFeedItem(feed)
		if likedFeed.FeedID != "" {
			likedFeeds = append(likedFeeds, likedFeed)
		}
	}

	return &UserLikesResponse{
		LikedFeeds: likedFeeds,
		Count:      len(likedFeeds),
		HasMore:    len(likedFeeds) > 0, // 简单判断，实际可能需要更复杂的逻辑
	}, nil
}

// extractFromDOM 从 DOM 元素提取数据
func (u *UserLikesAction) extractFromDOM(page *rod.Page) (*UserLikesResponse, error) {
	result := page.MustEval(`() => {
		const likedFeeds = [];

		// 查找笔记项的多种可能选择器
		const feedSelectors = [
			'.feeds-container .note-item',
			'.note-list .note-item',
			'.feeds-page .note-item',
			'[data-testid="note-item"]',
			'.note-card'
		];

		let feedItems = [];
		for (const selector of feedSelectors) {
			feedItems = document.querySelectorAll(selector);
			if (feedItems.length > 0) break;
		}

		for (const item of feedItems) {
			try {
				// 提取链接
				const linkElement = item.querySelector('a[href*="/explore/"]') ||
								  item.querySelector('a');
				const href = linkElement ? linkElement.href : '';

				// 提取标题
				const titleElement = item.querySelector('.title') ||
								   item.querySelector('.note-title') ||
								   item.querySelector('h3') ||
								   item.querySelector('[data-testid="note-title"]');
				const title = titleElement ? titleElement.textContent.trim() : '';

				// 提取作者信息
				const authorElement = item.querySelector('.author') ||
									item.querySelector('.username') ||
									item.querySelector('.user-name');
				const author = authorElement ? authorElement.textContent.trim() : '';

				// 提取 feed ID
				let feedId = '';
				if (href) {
					const match = href.match(/\/explore\/([a-f0-9]+)/);
					if (match) {
						feedId = match[1];
					}
				}

				if (feedId && title) {
					likedFeeds.push({
						feed_id: feedId,
						title: title,
						url: href,
						author: author,
						author_id: '',
						liked_time: new Date().toISOString()
					});
				}
			} catch (error) {
				console.warn('Error parsing feed item:', error);
			}
		}

		return JSON.stringify(likedFeeds);
	}`).String()

	var likedFeeds []LikedFeed
	if err := json.Unmarshal([]byte(result), &likedFeeds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DOM parsed data: %w", err)
	}

	return &UserLikesResponse{
		LikedFeeds: likedFeeds,
		Count:      len(likedFeeds),
		HasMore:    false,
	}, nil
}

// getStringValue 从 map 中按优先级获取字符串值
func (u *UserLikesAction) getStringValue(item map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := item[key].(string); ok {
			return value
		}
	}
	return ""
}

// extractAuthor 从 map 中提取作者信息
func (u *UserLikesAction) extractAuthor(item map[string]interface{}) string {
	if user, ok := item["user"].(map[string]interface{}); ok {
		if nickname, ok := user["nickname"].(string); ok {
			return nickname
		}
		if nickName, ok := user["nickName"].(string); ok {
			return nickName
		}
	}
	return ""
}

// parseFeedItem 解析单个笔记项
func (u *UserLikesAction) parseFeedItem(feed map[string]interface{}) LikedFeed {
	likedFeed := LikedFeed{}

	// 提取 ID
	if id, ok := feed["id"].(string); ok {
		likedFeed.FeedID = id
	} else if noteCard, ok := feed["noteCard"].(map[string]interface{}); ok {
		if id, ok := noteCard["id"].(string); ok {
			likedFeed.FeedID = id
		}
	}

	// 提取标题
	if noteCard, ok := feed["noteCard"].(map[string]interface{}); ok {
		if title, ok := noteCard["displayTitle"].(string); ok {
			likedFeed.Title = title
		}
	}

	// 提取作者信息
	if noteCard, ok := feed["noteCard"].(map[string]interface{}); ok {
		if user, ok := noteCard["user"].(map[string]interface{}); ok {
			if nickname, ok := user["nickname"].(string); ok {
				likedFeed.Author = nickname
			}
			if userId, ok := user["userId"].(string); ok {
				likedFeed.AuthorID = userId
			}
		}
	}

	// 构建 URL
	if likedFeed.FeedID != "" {
		likedFeed.URL = fmt.Sprintf("https://www.xiaohongshu.com/explore/%s", likedFeed.FeedID)
	}

	// 设置点赞时间
	likedFeed.LikedTime = time.Now().Format("2006-01-02 15:04:05")

	return likedFeed
}
