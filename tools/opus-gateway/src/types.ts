export type RunMode = "read-only" | "edit";

export interface RunRequestBody {
  prompt: string;
  cwd?: string;
  mode?: RunMode;
  allowBash?: boolean;
  maxTurns?: number;
  timeoutSec?: number;
}

export interface NormalizedRunRequest {
  prompt: string;
  cwd: string;
  mode: RunMode;
  allowBash: boolean;
  maxTurns: number;
  timeoutSec: number;
}

export interface ToolDenial {
  toolName: string;
  reason: string;
  input?: unknown;
}

export interface ToolSummary {
  counts: Record<string, number>;
  bashCommands: string[];
  denials: ToolDenial[];
}

export interface RunResult {
  status: "ok" | "error";
  model: string;
  finalText: string;
  toolSummary: ToolSummary;
  exitReason: "completed" | "max_turns" | "execution_error";
  durationMs: number;
}

export interface HealthResult {
  status: "ok" | "error";
  model: string;
  repoRoot: string;
  apiKeyConfigured: boolean;
  message?: string;
}

export interface LogEntry {
  timestamp: string;
  requestId: string;
  route: string;
  cwd?: string;
  mode?: RunMode;
  allowBash?: boolean;
  maxTurns?: number;
  timeoutSec?: number;
  promptPreview?: string;
  promptLength?: number;
  statusCode: number;
  durationMs?: number;
  resultStatus?: "ok" | "error";
  exitReason?: string;
  error?: string;
}
