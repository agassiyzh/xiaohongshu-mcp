package xiaohongshu

import (
	"context"
	"fmt"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

type NavigateAction struct {
	page *rod.Page
}

func NewNavigate(page *rod.Page) *NavigateAction {
	return &NavigateAction{page: page}
}

func (n *NavigateAction) ToExplorePage(ctx context.Context) error {
	page := n.page.Context(ctx)

	page.MustNavigate("https://www.xiaohongshu.com/explore").
		MustWaitLoad().
		MustElement(`div#app`)

	return nil
}

func (n *NavigateAction) ToProfilePage(ctx context.Context) error {
	page := n.page.Context(ctx)

	// First navigate to explore page
	if err := n.ToExplorePage(ctx); err != nil {
		return err
	}

	page.MustWaitStable()

	// Find and click the "我" channel link in sidebar
	profileLink := page.MustElement(`div.main-container li.user.side-bar-component a.link-wrapper span.channel`)
	profileLink.MustClick()

	// Wait for navigation to complete
	page.MustWaitLoad()

	return nil
}

func (n *NavigateAction) ToUserLikesPage(ctx context.Context) error {
	page := n.page.Context(ctx)

	// First navigate to likes page
	if err := n.ToProfilePage(ctx); err != nil {
		return err
	}

	page.MustWaitStable()

	likesTab, err := page.Element(`div.reds-tab-item.sub-tab-list span:contains("点赞")`)
	if err != nil {
		// 如果直接找不到，尝试使用 JavaScript 查找
		logrus.Info("Using JavaScript to find and click likes tab")
		result := page.MustEval(`() => {
				const tabs = document.querySelectorAll('div.reds-tab-item.sub-tab-list span');
				for (const tab of tabs) {
					if (tab.textContent === '点赞') {
						tab.click();
						return true;
					}
				}
				return false;
			}`).Bool()

		if !result {
			return fmt.Errorf("could not find likes tab")
		}
		logrus.Info("Successfully clicked likes tab using JavaScript")
	} else {
		likesTab.MustClick()
		logrus.Info("Successfully clicked likes tab")
	}

	// 等待导航完成
	page.MustWaitLoad()

	return nil
}
