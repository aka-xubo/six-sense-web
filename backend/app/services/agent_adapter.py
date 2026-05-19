"""AI Agent adapters for page analysis"""
import asyncio
import json
import re
from pathlib import Path
from typing import AsyncIterator, Optional
from abc import ABC, abstractmethod


class AgentAdapter(ABC):
    """Base class for AI Agent adapters"""

    @abstractmethod
    async def analyze_page(self, url: str, content: str) -> AsyncIterator[str]:
        """
        Analyze page content and yield streaming results

        Args:
            url: Page URL
            content: Page content (HTML)

        Yields:
            Streaming text output
        """
        pass

    def _build_prompt(self, url: str, content: str) -> str:
        """
        Build analysis prompt using template

        Args:
            url: Page URL
            content: Page content

        Returns:
            Formatted prompt
        """
        # Load prompt template
        template_path = Path(__file__).parent.parent.parent / "templates" / "insights_prompt_template.md"
        with open(template_path, 'r', encoding='utf-8') as f:
            template = f.read()

        # Build prompt
        prompt = f"""请分析以下网页内容，严格按照 insights_prompt_template.md 中的规范生成结构化的 insights。

网页 URL: {url}
网页内容: {content[:10000]}

要求：
1. summary 必须是20-40字的陈述句
2. type 必须从预定义列表中选择
3. keywords 必须恰好3个，无重复
4. 输出纯 JSON 格式，无其他文字

输出格式：
{{
  "summary": "...",
  "type": "...",
  "keywords": ["...", "...", "..."]
}}

模板规范：
{template}
"""
        return prompt


class ClaudeAdapter(AgentAdapter):
    """Claude Code CLI adapter"""

    async def analyze_page(self, url: str, content: str) -> AsyncIterator[str]:
        """
        Analyze page using Claude Code CLI

        Args:
            url: Page URL
            content: Page content

        Yields:
            Streaming text output
        """
        prompt = self._build_prompt(url, content)

        try:
            process = await asyncio.create_subprocess_exec(
                "claude",
                "--model", "sonnet",
                stdin=asyncio.subprocess.PIPE,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
        except FileNotFoundError as exc:
            raise RuntimeError("Claude Code CLI 不可用，请确认 claude 命令已正确安装") from exc

        # Write prompt
        process.stdin.write(prompt.encode('utf-8'))
        await process.stdin.drain()
        process.stdin.close()

        # Stream output
        async for line in process.stdout:
            text = line.decode('utf-8', errors='ignore')
            yield text

        return_code = await process.wait()
        if return_code != 0:
            stderr = await process.stderr.read()
            message = stderr.decode('utf-8', errors='ignore').strip()
            raise RuntimeError(message or f"Claude Code CLI exited with code {return_code}")


class CodexAdapter(AgentAdapter):
    """OpenAI Codex CLI adapter"""

    async def analyze_page(self, url: str, content: str) -> AsyncIterator[str]:
        """
        Analyze page using Codex CLI

        Args:
            url: Page URL
            content: Page content

        Yields:
            Streaming text output
        """
        prompt = self._build_prompt(url, content)

        try:
            process = await asyncio.create_subprocess_exec(
                "codex",
                "chat",
                stdin=asyncio.subprocess.PIPE,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
        except FileNotFoundError as exc:
            raise RuntimeError("Codex CLI 不可用，请确认 codex 命令已正确安装") from exc

        # Write prompt
        process.stdin.write(prompt.encode('utf-8'))
        await process.stdin.drain()
        process.stdin.close()

        # Stream output
        async for line in process.stdout:
            text = line.decode('utf-8', errors='ignore')
            yield text

        return_code = await process.wait()
        if return_code != 0:
            stderr = await process.stderr.read()
            message = stderr.decode('utf-8', errors='ignore').strip()
            raise RuntimeError(message or f"Codex CLI exited with code {return_code}")


def get_adapter(agent_name: str) -> AgentAdapter:
    """
    Get adapter for specified agent

    Args:
        agent_name: Agent name (claude, codex)

    Returns:
        Agent adapter instance

    Raises:
        ValueError: If agent not supported
    """
    adapters = {
        "claude": ClaudeAdapter,
        "codex": CodexAdapter
    }

    adapter_class = adapters.get(agent_name)
    if not adapter_class:
        raise ValueError(f"Unsupported agent: {agent_name}")

    return adapter_class()
