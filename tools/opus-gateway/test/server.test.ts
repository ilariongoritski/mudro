import test from "node:test";
import assert from "node:assert/strict";
import { once } from "node:events";
import type { AddressInfo } from "node:net";
import { createGatewayServer } from "../src/server.js";
import type { NormalizedRunRequest, RunResult } from "../src/types.js";

test("healthz reports missing api key", async () => {
  const server = createGatewayServer({
    env: {},
    repoRoot: "D:\\mudr\\mudro11"
  });

  server.listen(0, "127.0.0.1");
  await once(server, "listening");

  const { port } = server.address() as AddressInfo;
  const response = await fetch(`http://127.0.0.1:${port}/healthz`);
  const payload = await response.json();

  assert.equal(response.status, 503);
  assert.equal(payload.status, "error");
  assert.equal(payload.apiKeyConfigured, false);

  server.close();
});

test("v1/run forwards normalized request to runner", async () => {
  let capturedRequest: NormalizedRunRequest | undefined;

  const runner = async (_repoRoot: string, request: NormalizedRunRequest): Promise<RunResult> => {
    capturedRequest = request;

    return {
      status: "ok",
      model: "opus",
      finalText: "done",
      toolSummary: {
        counts: {
          Read: 1
        },
        bashCommands: [],
        denials: []
      },
      exitReason: "completed",
      durationMs: 5
    };
  };

  const server = createGatewayServer({
    env: {
      ANTHROPIC_API_KEY: "test-key"
    },
    repoRoot: "D:\\mudr\\mudro11",
    runner
  });

  server.listen(0, "127.0.0.1");
  await once(server, "listening");

  const { port } = server.address() as AddressInfo;
  const response = await fetch(`http://127.0.0.1:${port}/v1/run`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      prompt: "Inspect README",
      mode: "read-only"
    })
  });

  const payload = await response.json();

  assert.equal(response.status, 200);
  assert.equal(payload.status, "ok");
  assert.equal(capturedRequest?.prompt, "Inspect README");
  assert.equal(capturedRequest?.mode, "read-only");
  assert.equal(capturedRequest?.allowBash, false);

  server.close();
});
