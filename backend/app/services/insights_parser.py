"""Insights parser - extract structured insights from AI output"""
import json
import re
from typing import Optional


def parse_insights(text: str) -> Optional[dict]:
    """
    Parse insights from AI output text

    Args:
        text: AI output text (may contain JSON)

    Returns:
        Parsed insights dict or None if failed
    """
    # Try to find JSON in the text
    # Look for {...} pattern
    json_pattern = r'\{[^{}]*"summary"[^{}]*"type"[^{}]*"keywords"[^{}]*\}'
    matches = re.findall(json_pattern, text, re.DOTALL)

    if matches:
        # Try to parse the first match
        for match in matches:
            try:
                insights = json.loads(match)
                # Validate structure
                if _validate_insights(insights):
                    return insights
            except json.JSONDecodeError:
                continue

    # If no valid JSON found, try to parse the entire text
    try:
        insights = json.loads(text.strip())
        if _validate_insights(insights):
            return insights
    except json.JSONDecodeError:
        pass

    return None


def _validate_insights(insights: dict) -> bool:
    """
    Validate insights structure

    Args:
        insights: Insights dict

    Returns:
        True if valid, False otherwise
    """
    # Check required fields
    if not isinstance(insights, dict):
        return False

    if 'summary' not in insights or not isinstance(insights['summary'], str):
        return False

    if 'type' not in insights or not isinstance(insights['type'], str):
        return False

    if 'keywords' not in insights or not isinstance(insights['keywords'], list):
        return False

    # Check keywords count (should be exactly 3)
    if len(insights['keywords']) != 3:
        return False

    return True
