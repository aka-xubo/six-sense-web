"""AI Agent detection service"""
import shutil
import subprocess
from typing import Optional
from app.models import AgentInfo


# Supported AI Agents (MVP: Claude and Codex only)
# Cursor and Aider support planned for Phase 2
SUPPORTED_AGENTS = [
    {
        "name": "claude",
        "display_name": "Claude Code",
        "command": "claude",
        "version_flag": "--version"
    },
    {
        "name": "codex",
        "display_name": "OpenAI Codex",
        "command": "codex",
        "version_flag": "--version"
    }
]


class AgentDetector:
    """Detect available AI CLI agents"""

    def detect_agents(self) -> list[AgentInfo]:
        """
        Scan PATH and detect available AI CLI agents

        Returns:
            List of available agents with version info
        """
        available_agents = []

        for agent_config in SUPPORTED_AGENTS:
            # Check if command exists in PATH
            if shutil.which(agent_config["command"]):
                # Get version. Some CLIs leave a wrapper on PATH even when the
                # actual packaged binary is missing, so a failed version probe
                # means the agent is not usable.
                version = self._get_version(
                    agent_config["command"],
                    agent_config["version_flag"]
                )

                available_agents.append(
                    AgentInfo(
                        name=agent_config["name"],
                        display_name=agent_config["display_name"],
                        version=version,
                        available=version is not None
                    )
                )
            else:
                # Agent not available
                available_agents.append(
                    AgentInfo(
                        name=agent_config["name"],
                        display_name=agent_config["display_name"],
                        version=None,
                        available=False
                    )
                )

        return available_agents

    def _get_version(self, command: str, version_flag: str) -> Optional[str]:
        """
        Get version of a CLI command

        Args:
            command: Command name
            version_flag: Flag to get version (e.g., --version)

        Returns:
            Version string or None if failed
        """
        try:
            result = subprocess.run(
                [command, version_flag],
                capture_output=True,
                text=True,
                timeout=5
            )

            if result.returncode == 0:
                # Parse version from output
                output = result.stdout.strip()
                # Try to extract version number (e.g., "claude 2.1.0" -> "2.1.0")
                parts = output.split()
                for part in parts:
                    if part[0].isdigit():
                        return part
                return output
            else:
                return None
        except Exception:
            return None


# Global detector instance
agent_detector = AgentDetector()
