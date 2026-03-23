import path from "node:path";
import { fileURLToPath } from "node:url";
import { promises as fs } from "node:fs";

const DEFAULT_REPO_ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

export function getRepoRoot(): string {
  return process.env.OPUS_GATEWAY_REPO_ROOT
    ? path.resolve(process.env.OPUS_GATEWAY_REPO_ROOT)
    : DEFAULT_REPO_ROOT;
}

export function getPort(): number {
  const raw = process.env.OPUS_GATEWAY_PORT;

  if (!raw) {
    return 8788;
  }

  const port = Number(raw);
  return Number.isInteger(port) && port > 0 && port < 65536 ? port : 8788;
}

export function getDefaultModel(): string {
  return process.env.OPUS_GATEWAY_MODEL?.trim() || "opus";
}

export async function resolveWorkingDirectory(repoRoot: string, requestedCwd?: string): Promise<string> {
  const target = requestedCwd ? path.resolve(repoRoot, requestedCwd) : repoRoot;
  const repoReal = await fs.realpath(repoRoot);
  const targetReal = await resolveExistingPath(target, "cwd");

  if (!isWithinRoot(repoReal, targetReal)) {
    throw new Error(`cwd must stay inside repo root: ${repoReal}`);
  }

  const stats = await fs.stat(targetReal);
  if (!stats.isDirectory()) {
    throw new Error(`cwd must be a directory: ${targetReal}`);
  }

  return targetReal;
}

export function ensurePathInsideRepo(repoRoot: string, candidatePath: string, baseDirectory = repoRoot): void {
  const resolvedCandidate = path.resolve(baseDirectory, candidatePath);
  if (!isWithinRoot(repoRoot, resolvedCandidate)) {
    throw new Error(`path must stay inside repo root: ${repoRoot}`);
  }
}

export function getLogDirectory(repoRoot: string): string {
  return path.join(repoRoot, "var", "log", "opus-gateway");
}

function isWithinRoot(repoRoot: string, candidate: string): boolean {
  const root = normalizePath(repoRoot);
  const value = normalizePath(candidate);
  return value === root || value.startsWith(`${root}${path.sep}`);
}

function normalizePath(input: string): string {
  return path.resolve(input).replace(/[\\/]+$/u, "").toLowerCase();
}

async function resolveExistingPath(candidate: string, fieldName: string): Promise<string> {
  try {
    return await fs.realpath(candidate);
  } catch {
    throw new Error(`${fieldName} does not exist: ${candidate}`);
  }
}
