import { query } from "@anthropic-ai/claude-code";
import { buildAllowedTools, buildSystemPrompt, isAllowedBashCommand } from "./validation.js";
import { ensurePathInsideRepo, getDefaultModel } from "./paths.js";
import type { NormalizedRunRequest, RunResult, ToolSummary } from "./types.js";

type SdkMessage = {
  type: string;
  subtype?: string;
  model?: string;
  result?: string;
  duration_ms?: number;
  permission_denials?: Array<{
    tool_name?: string;
    tool_input?: unknown;
  }>;
};

export async function runClaudeTask(repoRoot: string, request: NormalizedRunRequest): Promise<RunResult> {
  const startedAt = Date.now();
  const allowedTools = buildAllowedTools(request.mode, request.allowBash);
  const allowedToolNames = new Set(allowedTools);
  const toolSummary: ToolSummary = {
    counts: {},
    bashCommands: [],
    denials: []
  };

  const abortController = new AbortController();
  const timeoutHandle = setTimeout(() => abortController.abort(), request.timeoutSec * 1000);

  let model = getDefaultModel();
  let finalText = "";
  let exitReason: RunResult["exitReason"] = "execution_error";
  let durationMs = 0;
  let sawResult = false;

  try {
    const stream = query({
      abortController,
      prompt: request.prompt,
      options: {
        allowedTools,
        canUseTool: async (toolName: string, input: unknown) => {
          try {
            validateToolUse(repoRoot, request, allowedToolNames, toolSummary, toolName, input);
            return {
              behavior: "allow" as const,
              updatedInput: input
            };
          } catch (error) {
            const message = error instanceof Error ? error.message : "tool denied";
            toolSummary.denials.push({
              toolName,
              reason: message,
              input
            });

            return {
              behavior: "deny" as const,
              message
            };
          }
        },
        cwd: request.cwd,
        maxTurns: request.maxTurns,
        model,
        permissionMode: request.mode === "edit" ? "acceptEdits" : "default",
        appendSystemPrompt: buildSystemPrompt(repoRoot, request)
      }
    });

    for await (const message of stream as AsyncIterable<SdkMessage>) {
      if (message.type === "system" && message.subtype === "init" && typeof message.model === "string") {
        model = message.model;
      }

      if (message.type !== "result") {
        continue;
      }

      sawResult = true;
      durationMs = typeof message.duration_ms === "number" ? message.duration_ms : durationMs;

      if (Array.isArray(message.permission_denials)) {
        for (const denial of message.permission_denials) {
          toolSummary.denials.push({
            toolName: denial.tool_name ?? "unknown",
            reason: "permission denied by Claude Code",
            input: denial.tool_input
          });
        }
      }

      if (message.subtype === "success") {
        finalText = typeof message.result === "string" ? message.result : "";
        exitReason = "completed";
        continue;
      }

      if (message.subtype === "error_max_turns") {
        exitReason = "max_turns";
        continue;
      }

      exitReason = "execution_error";
    }

    if (!sawResult) {
      throw new Error("Claude Code SDK returned no final result");
    }

    return {
      status: exitReason === "completed" ? "ok" : "error",
      model,
      finalText,
      toolSummary,
      exitReason,
      durationMs: durationMs || Date.now() - startedAt
    };
  } finally {
    clearTimeout(timeoutHandle);
  }
}

function validateToolUse(
  repoRoot: string,
  request: NormalizedRunRequest,
  allowedToolNames: Set<string>,
  toolSummary: ToolSummary,
  toolName: string,
  input: unknown
): void {
  if (!allowedToolNames.has(toolName)) {
    throw new Error(`tool is not enabled for this run: ${toolName}`);
  }

  switch (toolName) {
    case "Read": {
      const filePath = readStringField(input, "file_path");
      ensurePathInsideRepo(repoRoot, filePath, request.cwd);
      recordToolUse(toolSummary, toolName);
      return;
    }

    case "Glob": {
      const maybePath = optionalStringField(input, "path");
      if (maybePath) {
        ensurePathInsideRepo(repoRoot, maybePath, request.cwd);
      }
      recordToolUse(toolSummary, toolName);
      return;
    }

    case "Edit": {
      if (request.mode !== "edit") {
        throw new Error("edit mode is required for file edits");
      }

      const filePath = readStringField(input, "file_path");
      ensurePathInsideRepo(repoRoot, filePath, request.cwd);
      recordToolUse(toolSummary, toolName);
      return;
    }

    case "Bash": {
      if (!request.allowBash) {
        throw new Error("bash is disabled for this run");
      }

      const command = readStringField(input, "command").trim();
      if (!isAllowedBashCommand(command)) {
        throw new Error(`bash command is not allowlisted: ${command}`);
      }

      recordToolUse(toolSummary, toolName);
      toolSummary.bashCommands.push(command);
      return;
    }

    default:
      throw new Error(`unsupported tool: ${toolName}`);
  }
}

function recordToolUse(toolSummary: ToolSummary, toolName: string): void {
  toolSummary.counts[toolName] = (toolSummary.counts[toolName] ?? 0) + 1;
}

function readStringField(input: unknown, fieldName: string): string {
  if (!input || typeof input !== "object") {
    throw new Error(`tool input must be an object for ${fieldName}`);
  }

  const value = (input as Record<string, unknown>)[fieldName];
  if (typeof value !== "string" || !value.trim()) {
    throw new Error(`tool input must include string field: ${fieldName}`);
  }

  return value;
}

function optionalStringField(input: unknown, fieldName: string): string | undefined {
  if (!input || typeof input !== "object") {
    return undefined;
  }

  const value = (input as Record<string, unknown>)[fieldName];
  return typeof value === "string" && value.trim() ? value : undefined;
}
