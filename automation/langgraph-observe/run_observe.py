#!/usr/bin/env python3
"""Run the deterministic observe-only LangGraph with local sample data."""

from __future__ import annotations

import argparse
import uuid
from pathlib import Path

from observe_graph import build_graph


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--state-db", type=Path, default=Path("var/observe-state.sqlite"))
    parser.add_argument("--thread-id", default=None)
    args = parser.parse_args()

    graph, checkpointer = build_graph(args.state_db)
    thread_id = args.thread_id or f"observe-{uuid.uuid4()}"
    try:
        result = graph.invoke(
            {
                "run_id": thread_id,
                "observations": [
                    {"name": "example-healthz", "status": "ok", "detail": "sample only"},
                    {"name": "example-backup", "status": "healthy", "detail": "sample only"},
                ],
            },
            {"configurable": {"thread_id": thread_id}},
        )
        print(result["report"])
        return 0
    finally:
        checkpointer.conn.close()


if __name__ == "__main__":
    raise SystemExit(main())
