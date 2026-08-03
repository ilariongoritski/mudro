"""Observe-only LangGraph workflow for MUDRO autonomy preparation.

The graph is intentionally data-only: it never executes commands, reads secrets,
or mutates production state. A caller must collect observations separately and
pass them as input.
"""

from __future__ import annotations

import sqlite3
from pathlib import Path
from typing import Literal, TypedDict

from langgraph.checkpoint.sqlite import SqliteSaver
from langgraph.graph import END, START, StateGraph

Health = Literal["healthy", "degraded", "unknown"]


class Observation(TypedDict):
    name: str
    status: str
    detail: str


class ObserveState(TypedDict, total=False):
    run_id: str
    observations: list[Observation]
    health: Health
    recommended_action: str | None
    report: str


def classify_health(state: ObserveState) -> ObserveState:
    """Classify supplied read-only observations without side effects."""
    observations = state.get("observations", [])
    if not observations:
        return {
            "health": "unknown",
            "recommended_action": "Collect read-only observations before deciding on remediation.",
        }

    failed = [item["name"] for item in observations if item["status"] not in {"ok", "healthy"}]
    if failed:
        return {
            "health": "degraded",
            "recommended_action": f"Review failed observations: {', '.join(failed)}. No automatic remediation is allowed.",
        }
    return {"health": "healthy", "recommended_action": None}


def build_report(state: ObserveState) -> ObserveState:
    """Produce a stable owner-facing report from already-collected state."""
    lines = [
        "MUDRO OBSERVE-ONLY REPORT",
        f"Run: {state.get('run_id', 'unknown')}",
        f"Health: {state.get('health', 'unknown')}",
        "Observations:",
    ]
    for item in state.get("observations", []):
        lines.append(f"- {item['name']}: {item['status']} — {item['detail']}")
    recommendation = state.get("recommended_action")
    if recommendation:
        lines.append(f"Recommended action: {recommendation}")
    lines.append("Safety: observe-only; no command, deployment, restart, or write action was performed.")
    return {"report": "\n".join(lines)}


def build_graph(checkpoint_path: Path):
    """Build a persisted observe-only graph using a local SQLite checkpointer."""
    checkpoint_path.parent.mkdir(parents=True, exist_ok=True)
    # LangGraph may checkpoint from a worker thread; this local, process-scoped
    # SQLite database needs cross-thread access for that deterministic write.
    connection = sqlite3.connect(checkpoint_path, check_same_thread=False)
    checkpointer = SqliteSaver(connection)
    workflow = StateGraph(ObserveState)
    workflow.add_node("classify_health", classify_health)
    workflow.add_node("build_report", build_report)
    workflow.add_edge(START, "classify_health")
    workflow.add_edge("classify_health", "build_report")
    workflow.add_edge("build_report", END)
    return workflow.compile(checkpointer=checkpointer), checkpointer
