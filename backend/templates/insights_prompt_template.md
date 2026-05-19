# Insights 分析 Prompt 模板

## 目标
为浏览器历史页面生成结构化的内容摘要，帮助用户快速回忆页面内容。

## 输出格式（严格遵守）

```json
{
  "summary": "一句话摘要（20-40字，陈述句，说明页面的核心内容）",
  "type": "页面类型（从预定义列表中选择）",
  "keywords": ["关键词1", "关键词2", "关键词3"]
}
```

## 字段规范

### 1. summary（摘要）
**要求：**
- 长度：20-40个中文字符（或40-80个英文字符）
- 格式：陈述句，不使用"这是..."、"本文..."等开头
- 内容：说明页面的**核心主题**和**关键信息**
- 角度：从**用户为什么访问这个页面**的角度描述

**好的示例：**
- ✅ "GitHub PR讨论OAuth 2.0认证bug修复方案"
- ✅ "Stack Overflow解答React Hooks使用常见错误"
- ✅ "技术博客介绍Kubernetes集群部署最佳实践"

**不好的示例：**
- ❌ "这是一个关于编程的网页"（太泛，无具体信息）
- ❌ "本文详细介绍了..."（格式不对，太啰嗦）
- ❌ "很有用的技术文档"（主观评价，无实质内容）

### 2. type（页面类型）
**必须从以下列表中选择一个：**

| 类型 | 说明 | 示例 |
|------|------|------|
| `文档` | 技术文档、API文档、使用手册 | MDN、官方文档 |
| `文章` | 博客文章、技术博客、教程 | Medium、个人博客 |
| `问答` | Q&A网站的问题页面 | Stack Overflow、知乎 |
| `代码` | GitHub仓库、代码片段、Gist | GitHub、GitLab |
| `视频` | 视频网站的视频页面 | YouTube、B站 |
| `新闻` | 新闻报道、资讯 | 科技媒体、新闻网站 |
| `工具` | 在线工具、SaaS产品页面 | 在线编辑器、转换工具 |
| `社区` | 论坛、讨论区、社交媒体 | Reddit、Twitter、论坛 |
| `其他` | 无法归类到以上类型 | 特殊页面 |

**判断逻辑：**
1. 优先根据**域名**判断（github.com → 代码，stackoverflow.com → 问答）
2. 其次根据**页面结构**判断（有视频播放器 → 视频）
3. 最后根据**内容特征**判断（教程性质 → 文章）

### 3. keywords（关键词）
**要求：**
- 数量：**恰好3个**关键词
- 长度：每个关键词2-6个字符
- 内容：提取页面的**核心主题词**
- 优先级：技术术语 > 领域词汇 > 通用词汇

**选择原则：**
1. **技术术语优先**：OAuth、Kubernetes、React Hooks
2. **具体胜过抽象**：选"认证bug"而非"问题"
3. **避免停用词**：不要"的"、"和"、"是"等
4. **保持独立性**：3个关键词应该覆盖不同维度

**好的示例：**
- ✅ `["OAuth", "认证", "bug修复"]` - 覆盖技术、领域、动作
- ✅ `["Kubernetes", "集群部署", "最佳实践"]` - 技术、场景、方法
- ✅ `["React Hooks", "常见错误", "解决方案"]` - 技术、问题、答案

**不好的示例：**
- ❌ `["编程", "技术", "学习"]` - 太泛，无具体信息
- ❌ `["OAuth", "OAuth 2.0", "OAuth认证"]` - 重复，缺乏多样性
- ❌ `["这个", "那个", "问题"]` - 无意义词汇

## 分析流程

### 步骤1：快速扫描
- 查看页面标题
- 识别域名（判断页面类型）
- 浏览第一屏内容

### 步骤2：提取核心信息
- 主题是什么？
- 解决什么问题？
- 关键技术/概念是什么？

### 步骤3：生成结构化输出
- 用一句话总结核心内容（summary）
- 从预定义列表选择类型（type）
- 提取3个关键词（keywords）

### 步骤4：质量检查
- [ ] summary 长度在20-40字之间
- [ ] summary 是陈述句，无主观评价
- [ ] type 在预定义列表中
- [ ] keywords 恰好3个
- [ ] keywords 无重复、无停用词

## 特殊情况处理

### 1. 页面无法访问
```json
{
  "summary": "页面无法访问或已失效",
  "type": "其他",
  "keywords": ["无法访问", "失效", "错误"]
}
```

### 2. 页面内容过少
```json
{
  "summary": "页面内容不足，无法生成有效摘要",
  "type": "其他",
  "keywords": ["内容不足", "空页面", "无效"]
}
```

### 3. 非文本内容（纯图片/视频）
```json
{
  "summary": "多媒体内容页面，包含[图片/视频]资源",
  "type": "视频",  // 或 "其他"
  "keywords": ["多媒体", "视觉内容", "资源"]
}
```

### 4. 登录墙/付费墙
```json
{
  "summary": "需要登录或付费才能访问的内容",
  "type": "其他",
  "keywords": ["需要登录", "付费内容", "受限访问"]
}
```

## 质量标准

### 优秀（90分+）
- summary 精准概括核心内容，用户一眼就能回忆起��面
- type 分类准确
- keywords 覆盖主题、技术、场景三个维度

### 良好（70-89分）
- summary 基本准确，但可能略显泛化
- type 分类正确
- keywords 有2个以上有价值

### 及格（60-69分）
- summary 能说明页面大致内容
- type 分类基本合理
- keywords 至少1个有价值

### 不及格（<60分）
- summary 过于泛化或错误
- type 分类错误
- keywords 无意义或重复

## 示例

### 示例1：技术文档
**URL:** https://kubernetes.io/docs/concepts/workloads/pods/

**分析：**
- 域名：kubernetes.io → 官方文档
- 标题：Pods | Kubernetes
- 内容：介绍 Kubernetes Pod 概念

**输出：**
```json
{
  "summary": "Kubernetes官方文档介绍Pod概念和使用方法",
  "type": "文档",
  "keywords": ["Kubernetes", "器编排"]
}
```

### 示例2：问答页面
**URL:** https://stackoverflow.com/questions/12345/how-to-use-react-hooks

**分析：**
- 域名：stackoverflow.com → 问答网站
- 标题：How to use React Hooks?
- 内容：React Hooks 使用问题及解答

**输出：**
```json
{
  "summary": "Stack Overflow讨论React Hooks基础用法和常见问题",
  "type": "问答",
  "keywords": ["React Hooks", "使用方法", "问题解答"]
}
```

### 示例3：GitHub PR
**URL:** https://github.com/user/repo/pull/123

**分析：**
- 域名：github.com → 代码托管
- 路径：/pull/ → Pull Request
- 标题：Fix authentication bug with OAuth 2.0

**输出：**
```json
{
  "summary": "GitHub PR修复OAuth 2.0认证相关bug",
  "type": "代码",
  "keywords": ["OAuth", "认证bug", "PR修复"]
}
```

### 示例4：技术博客
**URL:** https://blog.example.com/kubernetes-best-practices

**分析：**
- 域名：blog.* → 博客
- 标题：Kubernetes Deployment Best Practices
- 内容：Kubernetes 部署最佳实践教程

**输出：**
```json
{
  "summary": "技术博客分享Kubernetes生产环境部署最佳实践",
  "type": "文章",
  "keywords": ["Kubernetes", "部署实践", "生产环境"]
}
```

## 使用此模板

当 AI Agent 分析页面时，应该：
1. 使用 WebFetch 获取页面内容
2. 将页面内容和此模板一起发送给 AI
3. 要求 AI 严格按照模板生成 JSON 输出
4. 验证输出格式是否符合规范
5. 保存到 insights_cache.json

**Prompt 示例：**
```
请分析以下网页内容，严格按照 insights_prompt_template.md 中的规范生成结构化的 insights。

网页 URL: {url}
网页内容: {content}

要求：
1. summary 必须是20-40字的陈述句
2. type 必须从预定义列表中选择
3. keywords 必须恰好3个，无重复
4. 输出纯 JSON 格式，无其他文字

输出格式：
{
  "summary": "...",
  "type": "...",
  "keywords": ["...", "...", "..."]
}
```
