import test from "node:test";
import assert from "node:assert/strict";
import { getDefaultModel } from "../src/paths.js";

test("getDefaultModel prefers explicit gateway override", () => {
  const previousGatewayModel = process.env.OPUS_GATEWAY_MODEL;
  const previousAnthropicModel = process.env.ANTHROPIC_MODEL;

  process.env.OPUS_GATEWAY_MODEL = "gateway-model";
  process.env.ANTHROPIC_MODEL = "anthropic-model";

  assert.equal(getDefaultModel(), "gateway-model");

  restoreEnv("OPUS_GATEWAY_MODEL", previousGatewayModel);
  restoreEnv("ANTHROPIC_MODEL", previousAnthropicModel);
});

test("getDefaultModel falls back to Anthropic model and pinned snapshot", () => {
  const previousGatewayModel = process.env.OPUS_GATEWAY_MODEL;
  const previousAnthropicModel = process.env.ANTHROPIC_MODEL;

  delete process.env.OPUS_GATEWAY_MODEL;
  process.env.ANTHROPIC_MODEL = "anthropic-model";
  assert.equal(getDefaultModel(), "anthropic-model");

  delete process.env.ANTHROPIC_MODEL;
  assert.equal(getDefaultModel(), "claude-opus-4-1-20250805");

  restoreEnv("OPUS_GATEWAY_MODEL", previousGatewayModel);
  restoreEnv("ANTHROPIC_MODEL", previousAnthropicModel);
});

function restoreEnv(name: "OPUS_GATEWAY_MODEL" | "ANTHROPIC_MODEL", value: string | undefined): void {
  if (value === undefined) {
    delete process.env[name];
    return;
  }

  process.env[name] = value;
}
