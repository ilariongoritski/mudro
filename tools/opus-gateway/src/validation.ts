import type { NormalizedRunRequest, RunMode, RunRequestBody } from "./types.js";

const MAX_TURNS_DEFAULT = 10;
const TIMEOUT_SEC_DEFAULT = 600;

export function normalizeRunRequest(body: unknown, cwd: string): NormalizedRunRequest {
  if (!isObject(body)) {
    throw new Error("request body must be a JSON object");
  }

  const prompt = typeof body.prompt === "string" ? body.prompt.trim() : "";
  if (!prompt) {
    throw new Error("prompt is required");
  }

  const mode = normalizeMode(body.mode);
  const allowBash = body.allowBash === true;
  const maxTurns = normalizeInteger(body.maxTurns, MAX_TURNS_DEFAULT, 1, 50, "maxTurns");
  const timeoutSec = normalizeInteger(body.timeoutSec, TIMEOUT_SEC_DEFAULT, 5, 1800, "timeoutSec");

  return {
    prompt,
    cwd,
    mode,
    allowBash,
    maxTurns,
    timeoutSec
  };
}

export function getRequestedCwd(body: unknown): string | undefined {
  if (!isObject(body) || body.cwd === undefined) {
    return undefined;
  }

  if (typeof body.cwd !== "string" || !body.cwd.trim()) {
    throw new Error("cwd must be a non-empty string when provided");
  }

  return body.cwd.trim();
}

export function buildAllowedTools(mode: RunMode, allowBash: boolean): string[] {
  const tools = mode === "edit" ? ["Read", "Glob", "Edit"] : ["Read", "Glob"];

  if (allowBash) {
    tools.push("Bash");
  }

  return tools;
}

export function isAllowedBashCommand(command: string): boolean {
  const normalized = command.trim();

  if (!normalized) {
    return false;
  }

  if (/[;&|><`]/u.test(normalized) || /\$\(/u.test(normalized) || /[\r\n]/u.test(normalized)) {
    return false;
  }

  if (
    normalized === "rg" ||
    normalized.startsWith("rg ") ||
    normalized === "go test" ||
    normalized === "go test ./..." ||
    normalized.startsWith("go test ") ||
    normalized === "npm test" ||
    normalized === "npm run lint" ||
    normalized === "npm run build" ||
    normalized === "git status" ||
    normalized.startsWith("git status ") ||
    normalized === "git diff" ||
    normalized.startsWith("git diff ")
  ) {
    return true;
  }

  return false;
}

export function buildSystemPrompt(repoRoot: string, request: NormalizedRunRequest): string {
  const bashPolicy = request.allowBash
    ? "Bash is available, but only for the allowlisted commands enforced by the gateway."
    : "Bash is disabled for this run.";

  const editPolicy =
    request.mode === "edit"
      ? "You may edit files inside the repository root when it materially improves the result."
      : "Do not attempt to edit files or suggest edits through tools.";

  return [
    "You are running inside a local MUDRO sidecar gateway.",
    `Repository root: ${repoRoot}`,
    `Working directory: ${request.cwd}`,
    "Stay inside the repository root.",
    "Prefer concise, implementation-focused answers.",
    bashPolicy,
    editPolicy,
    "Do not attempt network access through shell commands.",
    "Summarize exactly what you changed or inspected."
  ].join(" ");
}

function normalizeMode(raw: unknown): RunMode {
  if (raw === undefined) {
    return "read-only";
  }

  if (raw === "read-only" || raw === "edit") {
    return raw;
  }

  throw new Error('mode must be either "read-only" or "edit"');
}

function normalizeInteger(
  raw: unknown,
  fallback: number,
  min: number,
  max: number,
  fieldName: keyof RunRequestBody
): number {
  if (raw === undefined) {
    return fallback;
  }

  if (!Number.isInteger(raw) || raw < min || raw > max) {
    throw new Error(`${fieldName} must be an integer between ${min} and ${max}`);
  }

  return raw;
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
