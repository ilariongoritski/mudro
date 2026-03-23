import path from "node:path";
import { promises as fs } from "node:fs";
import { getLogDirectory } from "./paths.js";
import type { LogEntry } from "./types.js";

export async function appendGatewayLog(repoRoot: string, entry: LogEntry): Promise<void> {
  try {
    const logDirectory = getLogDirectory(repoRoot);
    await fs.mkdir(logDirectory, { recursive: true });

    const fileName = `${entry.timestamp.slice(0, 10)}.jsonl`;
    const line = `${JSON.stringify(entry)}\n`;

    await fs.appendFile(path.join(logDirectory, fileName), line, "utf8");
  } catch {
    // Logging failures must not fail the request path.
  }
}
