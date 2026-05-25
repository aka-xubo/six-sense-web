# Six Sense Web - Claude Code 协作规范

## 文档管理规范

### Obsidian 集成

所有的设计文档、计划、规范文档都应该写入 Obsidian vault 中对应的项目文件夹。

**⚠️ 重要原则: 读取优先 (Read First)**

在创建任何新文档之前，**必须先检查** Obsidian vault 中是否已有相关文档：

1. **定位 Obsidian Vault**
   ```bash
   # 查找 .obsidian 配置目录
   find ~ -type d -name ".obsidian" 2>/dev/null | head -5
   ```

2. **搜索现有文档**
   ```bash
   # 在项目目录下搜索相关文档
   find /path/to/vault/Projects/six-sense-web -type f -name "*.md"
   ```

3. **读取现有文档**
   - 如果找到相关文档，**必须先读取**内容
   - 理解现有设计/计划后再决定下一步行动
   - 避免重复创建或重复工作

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
 frontmatter:
```yaml
---
title: 文档标题
date: YYYY-MM-DD
tags: [six-sense, design, web]
status: draft|in-progress|completed
---
```

### 文档创建流程

**必须按顺序执行:**

1. **检查 Obsidian vault 位置**
   ```bash
   find ~ -type d -name ".obsidian" 2>/dev/null
   ```

2. **搜索现有文档**
   ```bash
   find /path/to/vault/Projects/six-sense-web -type f -name "*.md"
   ```

3. **读取相关文档**（如果存在）
   - 使用 `Read` 工具读取文档内容
   - 理解现有设计和决策
   - 避免重复工作

4. **创建新文档**（如果需要）
   - 如果项目文件夹不存在,先创建: `Projects/six-sense-web/`
   - 根据文档类型创建到对应子文件夹
   - 使用 Obsidian 语法(wikilinks, callouts 等)
   - 使用 `Write` 工具直接写入文件（不需要 obsidian-cli）

5. **更新文档状态**
   - 完成后更新 frontmatter 中的 status 为 completed

## 项目信息

- **项目名称**: Six Sense Web
- **项目类型**: 浏览器标签页管理工具 (Web 版本)
- **技术栈**: 
  - 后端: Go + net/http + SQLite
  - 前端: React + TypeScript + Vite + TailwindCSS
  - AI 集成: 本地 CLI 进程调用 (Claude/Cursor/Codex/Aider)
- **部署方式**: 纯本地应用 (localhost)

## 开发原则

1. **Local-First**: 所有数据和 AI 处理都在本地
2. **按需加载**: 同步和分析都由用户主动触发
3. **流式体验**: AI 分析结果实时显示,可中断
4. **渐进增强**: MVP 先实现核心功能,逐步迭代

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

Tradeoff: These guidelines bias toward caution over speed. For trivial tasks, use judgment.

1. Think Before Coding
Don't assume. Don't hide confusion. Surface tradeoffs.

Before implementing:

State your assumptions explicitly. If uncertain, ask.
If multiple interpretations exist, present them - don't pick silently.
If a simpler approach exists, say so. Push back when warranted.
If something is unclear, stop. Name what's confusing. Ask.
2. Simplicity First
Minimum code that solves the problem. Nothing speculative.

No features beyond what was asked.
No abstractions for single-use code.
No "flexibility" or "configurability" that wasn't requested.
No error handling for impossible scenarios.
If you write 200 lines and it could be 50, rewrite it.
Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

3. Surgical Changes
Touch only what you must. Clean up only your own mess.

When editing existing code:

Don't "improve" adjacent code, comments, or formatting.
Don't refactor things that aren't broken.
Match existing style, even if you'd do it differently.
If you notice unrelated dead code, mention it - don't delete it.
When your changes create orphans:

Remove imports/variables/functions that YOUR changes made unused.
Don't remove pre-existing dead code unless asked.
The test: Every changed line should trace directly to the user's request.

4. Goal-Driven Execution
Define success criteria. Loop until verified.

Transform tasks into verifiable goals:

"Add validation" → "Write tests for invalid inputs, then make them pass"
"Fix the bug" → "Write a test that reproduces it, then make it pass"
"Refactor X" → "Ensure tests pass before and after"
For multi-step tasks, state a brief plan:

1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

These guidelines are working if: fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.

## 参考资料

- **1.0 版本仓库**: https://github.com/aka-xubo/six-sense
- **设计文档**: 查看 Obsidian vault 中的 `Projects/six-sense-web/designs/`
- **实现计划**: 查看 Obsidian vault 中的 `Projects/six-sense-web/plans/`
