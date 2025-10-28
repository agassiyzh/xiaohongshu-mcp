# 获取用户点赞笔记功能文档

## 功能概述

新增了一个 MCP 工具 `get_user_liked_feeds`，用于获取当前登录用户所有点赞的笔记列表。

## MCP 工具使用

### 工具名称
`get_user_liked_feeds`

### 描述
获取当前登录用户所有点赞的笔记列表，返回笔记标题、链接、作者等详细信息。

### 参数
无需任何参数 - 工具会自动获取当前登录用户的点赞笔记。

### 返回数据结构
```json
{
  "liked_feeds": [
    {
      "feed_id": "65f123abc456def789",
      "title": "超美味的家常菜谱分享",
      "url": "https://www.xiaohongshu.com/explore/65f123abc456def789",
      "author": "美食达人小王",
      "author_id": "5b8c9d0e1f2a3b4c",
      "liked_time": "2024-01-15 14:30:00"
    }
  ],
  "count": 1,
  "has_more": false
}
```

### 使用示例

#### Claude Desktop 配置
```json
{
  "mcpServers": {
    "xiaohongshu-mcp": {
      "url": "http://localhost:18060/mcp"
    }
  }
}
```

#### AI 助手调用示例
```
请帮我获取我点赞的所有笔记列表
```

AI 助手会自动调用 `get_user_liked_feeds` 工具并返回结果。

## REST API 使用

### 端点
`GET /api/v1/user/liked-feeds`

### 请求头
```
Content-Type: application/json
```

### 响应
```json
{
  "success": true,
  "data": {
    "liked_feeds": [
      {
        "feed_id": "65f123abc456def789",
        "title": "超美味的家常菜谱分享",
        "url": "https://www.xiaohongshu.com/explore/65f123abc456def789",
        "author": "美食达人小王",
        "author_id": "5b8c9d0e1f2a3b4c",
        "liked_time": "2024-01-15 14:30:00"
      }
    ],
    "count": 1,
    "has_more": false
  },
  "message": "获取用户点赞笔记成功"
}
```

### cURL 示例
```bash
curl -X GET http://localhost:18060/api/v1/user/liked-feeds
```

## 技术实现细节

### 实现原理
1. **页面导航**: 自动导航到用户个人主页的点赞标签页
2. **数据提取**: 通过多种策略提取点赞笔记数据：
   - 从 `__INITIAL_STATE__` 获取数据
   - 从页面 DOM 结构解析
   - 滚动加载更多内容
3. **数据转换**: 将原始数据转换为标准化的响应格式

### 关键特性
- **自动登录检查**: 确保用户已登录
- **多重提取策略**: 确保在不同页面版本下都能正常工作
- **错误处理**: 完善的错误处理和恢复机制
- **性能优化**: 智能等待和页面稳定性检查

### 依赖组件
- `xiaohongshu/user_likes`: 核心业务逻辑
- `xiaohongshu/navigate`: 页面导航功能
- `go-rod`: 浏览器自动化

## 注意事项

### 使用前提
1. **必须已登录**: 使用前需要确保用户已登录小红书
2. **网络连接**: 需要稳定的网络连接访问小红书
3. **浏览器环境**: 需要支持的无头浏览器环境

### 限制说明
1. **数据时效性**: 获取的是当前时刻的点赞状态
2. **分页加载**: 目前只加载第一页内容，后续版本会支持分页
3. **隐私设置**: 受小红书隐私设置影响

### 故障排除

#### 常见错误
1. **未登录错误**
   ```
   获取用户点赞笔记失败: 用户未登录
   ```
   解决：先调用登录工具完成登录

2. **页面结构变化**
   ```
   无法从页面提取点赞笔记数据
   ```
   解决：等待更新或联系开发者

3. **网络超时**
   ```
   操作超时
   ```
   解决：检查网络连接，重试操作

## 开发信息

### 文件结构
```
xiaohongshu/user_likes/
├── likes.go          # 核心实现
└── likes_test.go     # 单元测试

# 相关修改文件
├── service.go        # 服务层接口
├── mcp_handlers.go   # MCP 处理器
├── mcp_server.go     # MCP 工具注册
├── handlers_api.go   # REST API 处理器
└── routes.go         # 路由配置
```

### 版本历史
- **v1.0.0**: 初始版本，支持获取用户点赞笔记

### 测试覆盖
- 单元测试：核心逻辑测试
- 集成测试：真实环境测试（需要登录状态）
- 基准测试：性能测试

## 相关功能

- `get_my_profile`: 获取当前用户基本信息
- `like_feed`: 点赞指定笔记
- `unlike_feed`: 取消点赞指定笔记
- `get_feed_detail`: 获取笔记详情

## 支持与反馈

如有问题或建议，请通过以下方式联系：
- GitHub Issues
- 项目文档
- 开发者社区