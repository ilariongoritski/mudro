import http, { type IncomingMessage, type ServerResponse } from "node:http";
import { randomUUID } from "node:crypto";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { appendGatewayLog } from "./logging.js";
import { getDefaultModel, getPort, getRepoRoot, resolveWorkingDirectory } from "./paths.js";
import { runClaudeTask } from "./runner.js";
import { getRequestedCwd, normalizeRunRequest } from "./validation.js";
import type { HealthResult, LogEntry, NormalizedRunRequest, RunResult } from "./types.js";

type Runner = (repoRoot: string, request: NormalizedRunRequest) => Promise<RunResult>;

interface GatewayOptions {
  env?: NodeJS.ProcessEnv;
  repoRoot?: string;
  runner?: Runner;
}

export function createGatewayServer(options: GatewayOptions = {}): http.Server {
  const env = options.env ?? process.env;
  const repoRoot = options.repoRoot ?? getRepoRoot();
  const runner = options.runner ?? runClaudeTask;

  let activeRequest = false;

  return http.createServer(async (req, res) => {
    const startedAt = Date.now();
    const requestId = randomUUID();

    try {
      if (req.method === "GET" && req.url === "/healthz") {
        return sendJson(res, env.ANTHROPIC_API_KEY ? 200 : 503, buildHealth(repoRoot, env));
      }

      if (req.method === "POST" && req.url === "/v1/run") {
        if (!env.ANTHROPIC_API_KEY) {
          await logRequest(repoRoot, {
            requestId,
            route: "/v1/run",
            statusCode: 503,
            error: "ANTHROPIC_API_KEY is not set"
          }, startedAt);

          return sendJson(res, 503, {
            status: "error",
            message: "ANTHROPIC_API_KEY is not set"
          });
        }

        if (activeRequest) {
          await logRequest(repoRoot, {
            requestId,
            route: "/v1/run",
            statusCode: 409,
            error: "gateway is busy"
          }, startedAt);

          return sendJson(res, 409, {
            status: "error",
            message: "gateway is busy, wait for the current run to finish"
          });
        }

        activeRequest = true;

        try {
          const body = await readJsonBody(req);
          const resolvedCwd = await resolveWorkingDirectory(repoRoot, getRequestedCwd(body));
          const runRequest = normalizeRunRequest(body, resolvedCwd);
          const result = await runner(repoRoot, runRequest);

          await logRequest(repoRoot, {
            requestId,
            route: "/v1/run",
            cwd: runRequest.cwd,
            mode: runRequest.mode,
            allowBash: runRequest.allowBash,
            maxTurns: runRequest.maxTurns,
            timeoutSec: runRequest.timeoutSec,
            promptPreview: runRequest.prompt.slice(0, 160),
            promptLength: runRequest.prompt.length,
            statusCode: result.status === "ok" ? 200 : 500,
            resultStatus: result.status,
            exitReason: result.exitReason
          }, startedAt, result.durationMs);

          return sendJson(res, result.status === "ok" ? 200 : 500, result);
        } catch (error) {
          const message = error instanceof Error ? error.message : "unexpected error";
          const statusCode = isClientError(message) ? 422 : 500;

          await logRequest(repoRoot, {
            requestId,
            route: "/v1/run",
            statusCode,
            error: message
          }, startedAt);

          return sendJson(res, statusCode, {
            status: "error",
            message
          });
        } finally {
          activeRequest = false;
        }
      }

      return sendJson(res, 404, {
        status: "error",
        message: "not found"
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : "unexpected server error";
      await logRequest(repoRoot, {
        requestId,
        route: req.url ?? "unknown",
        statusCode: 500,
        error: message
      }, startedAt);

      return sendJson(res, 500, {
        status: "error",
        message
      });
    }
  });
}

function buildHealth(repoRoot: string, env: NodeJS.ProcessEnv): HealthResult {
  const apiKeyConfigured = Boolean(env.ANTHROPIC_API_KEY);

  return {
    status: apiKeyConfigured ? "ok" : "error",
    model: getDefaultModel(),
    repoRoot,
    apiKeyConfigured,
    message: apiKeyConfigured ? "ok" : "ANTHROPIC_API_KEY is not set"
  };
}

async function logRequest(
  repoRoot: string,
  entry: Omit<LogEntry, "timestamp" | "durationMs">,
  startedAt: number,
  durationOverride?: number
): Promise<void> {
  await appendGatewayLog(repoRoot, {
    timestamp: new Date().toISOString(),
    durationMs: durationOverride ?? Date.now() - startedAt,
    ...entry
  });
}

async function readJsonBody(req: IncomingMessage): Promise<unknown> {
  const chunks: Buffer[] = [];
  let totalBytes = 0;

  for await (const chunk of req) {
    const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk);
    totalBytes += buffer.byteLength;

    if (totalBytes > 1024 * 1024) {
      throw new Error("request body is too large");
    }

    chunks.push(buffer);
  }

  if (chunks.length === 0) {
    throw new Error("request body is required");
  }

  try {
    return JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } catch {
    throw new Error("request body must be valid JSON");
  }
}

function sendJson(res: ServerResponse, statusCode: number, body: unknown): void {
  res.statusCode = statusCode;
  res.setHeader("Content-Type", "application/json; charset=utf-8");
  res.end(JSON.stringify(body, null, 2));
}

function isClientError(message: string): boolean {
  return [
    "prompt is required",
    "request body",
    "mode must",
    "cwd must",
    "does not exist",
    "must be a directory",
    "must stay inside repo root",
    "maxTurns must",
    "timeoutSec must"
  ].some((marker) => message.includes(marker));
}

const isMainModule =
  typeof process.argv[1] === "string" &&
  path.resolve(process.argv[1]) === fileURLToPath(import.meta.url);

if (isMainModule) {
  const repoRoot = getRepoRoot();
  const port = getPort();
  const server = createGatewayServer({ repoRoot });

  server.listen(port, "127.0.0.1", () => {
    process.stdout.write(
      `Opus gateway listening on http://127.0.0.1:${port} for repo ${repoRoot}\n`
    );
  });
}
