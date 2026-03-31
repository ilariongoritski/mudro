import test from "node:test";
import assert from "node:assert/strict";
import { buildAllowedTools, buildSystemPrompt, isAllowedBashCommand, normalizeRunRequest } from "../src/validation.js";

test("normalizeRunRequest applies defaults", () => {
  const request = normalizeRunRequest(
    {
      prompt: "  inspect services/feed-api  "
    },
    "E:\\mudr\\mudro11"
  );

  assert.deepEqual(request, {
    prompt: "inspect services/feed-api",
    cwd: "E:\\mudr\\mudro11",
    mode: "read-only",
    allowBash: false,
    maxTurns: 10,
    timeoutSec: 600
  });
});

test("normalizeRunRequest rejects invalid mode", () => {
  assert.throws(
    () =>
      normalizeRunRequest(
        {
          prompt: "inspect",
          mode: "write"
        },
        "E:\\mudr\\mudro11"
      ),
    /mode must be either "read-only" or "edit"/u
  );
});

test("buildAllowedTools enables edit and bash only when requested", () => {
  assert.deepEqual(buildAllowedTools("read-only", false), ["Read", "Glob"]);
  assert.deepEqual(buildAllowedTools("edit", true), ["Read", "Glob", "Edit", "Write", "Bash"]);
});

test("isAllowedBashCommand accepts only the allowlist", () => {
  assert.equal(isAllowedBashCommand("git status --short"), true);
  assert.equal(isAllowedBashCommand("rg gateway src"), true);
  assert.equal(isAllowedBashCommand("go test ./services/..."), true);
  assert.equal(isAllowedBashCommand("git diff && go test ./..."), false);
  assert.equal(isAllowedBashCommand("rm -rf ."), false);
});

test("buildSystemPrompt reflects mode and bash policy", () => {
  const prompt = buildSystemPrompt("E:\\mudr\\mudro11", {
    prompt: "inspect",
    cwd: "E:\\mudr\\mudro11",
    mode: "edit",
    allowBash: true,
    maxTurns: 10,
    timeoutSec: 600
  });

  assert.match(prompt, /Repository root: E:\\mudr\\mudro11/u);
  assert.match(prompt, /Bash is available/u);
  assert.match(prompt, /You may edit files/u);
});
