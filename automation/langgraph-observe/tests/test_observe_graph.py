from pathlib import Path

from observe_graph import build_graph


def invoke(graph, thread_id: str, observations: list[dict[str, str]]):
    return graph.invoke(
        {"run_id": thread_id, "observations": observations},
        {"configurable": {"thread_id": thread_id}},
    )


def test_healthy_observations_are_reported_and_persisted(tmp_path: Path):
    graph, checkpointer = build_graph(tmp_path / "state.sqlite")
    try:
        result = invoke(
            graph,
            "healthy-run",
            [{"name": "healthz", "status": "ok", "detail": "200"}],
        )
        assert result["health"] == "healthy"
        assert result["recommended_action"] is None
        assert "MUDRO OBSERVE-ONLY REPORT" in result["report"]
        assert "no command, deployment, restart, or write action was performed" in result["report"]
        assert checkpointer.get_tuple({"configurable": {"thread_id": "healthy-run"}}) is not None
    finally:
        checkpointer.conn.close()


def test_failed_observation_never_recommends_automatic_remediation(tmp_path: Path):
    graph, checkpointer = build_graph(tmp_path / "state.sqlite")
    try:
        result = invoke(
            graph,
            "degraded-run",
            [{"name": "backup", "status": "failed", "detail": "checksum mismatch"}],
        )
        assert result["health"] == "degraded"
        assert "backup" in result["recommended_action"]
        assert "No automatic remediation" in result["recommended_action"]
    finally:
        checkpointer.conn.close()


def test_empty_observations_are_unknown(tmp_path: Path):
    graph, checkpointer = build_graph(tmp_path / "state.sqlite")
    try:
        result = invoke(graph, "empty-run", [])
        assert result["health"] == "unknown"
        assert "Collect read-only observations" in result["recommended_action"]
    finally:
        checkpointer.conn.close()
