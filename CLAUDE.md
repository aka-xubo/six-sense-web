# Six Sense Web - Claude Code 协作规范

## 文档管理规范

### Obsidian 集成

所有的设计文档、计划、规范文档都应该写入 Obsidian vault 中对应的项目文件夹。

**文档存放路径规则:**
- 设计文档: `Obsidian Vault/Projects/six-sense-web/designs/`
- 实现计划: `Obsidian Vault/Projects/six-sense-web/plans/`
- 会议记录: `Obsidian Vault/Projects/six-sense-web/meetings/`
- 技术决策: `Obsidian Vault/Projects/six-sense-web/decisions/`

**文档命名规范:**
- 使用日期前缀: `YYYY-MM-DD-<topic>.md`
- 使用小写字母和连字符
- 示例: `2026-05-18-web-version-design.md`

**Obsidian 元数据:**
每个文档应包含 frontmatter:
```yaml
---
title: 文档标题
date: YYYY-MM-DD
tags: [six-sense, design, web]
status: draft|in-progress|completed
---
```

### 文档创建流程

1. 如果项目文件夹不存在,先创建: `Projects/six-sense-web/`
2. 根据文档类型创建到对应子文件夹
3. 使用 Obsidian 语法(wikilinks, callouts 等)
4. 完成后更新 status 为 completed

## 项目信息

- **项目名称**: Six Sense Web
- **项目类型**: 浏览器标签页管理工具 (Web 版本)
- **技术栈**: 
  - 后端: Python + FastAPI + SQLite
  - 前端: React + TypeScript + Vite + TailwindCSS
  - AI 集成: 本地 CLI 进程调用 (Claude/Cursor/Codex/Aider)
- **部署方式**: 纯本地应用 (localhost)

## 开发原则

1. **Local-First**: 所有数据和 AI 处理都在本地
2. **按需加载**: 同步和分析都由用户主动触发
3. **流式体验**: AI 分析结果实时显示,可中断
4. **渐进增强**: MVP 先实现核心功能,逐步迭代

## 代码复用

- 复用现有 Python 脚本的核心逻辑 (`scripts/six_sense.py`, `insights_analyzer.py`)
- 复用现有的 prompt 模板 (`templates/insights_prompt_template.md`)
- 保持与原有脚本版本的数据格式兼容性(通过导入/导出)
