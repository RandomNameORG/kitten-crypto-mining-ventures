import type { ActionRequest, Snapshot } from "./types";

const ARRAY_KEYS: ReadonlyArray<keyof Snapshot> = [
  "rooms",
  "gpus",
  "gpu_defs",
  "skills",
  "mercs",
  "merc_defs",
  "log",
];

function normalize(data: Snapshot): Snapshot {
  for (const key of ARRAY_KEYS) {
    if (!Array.isArray(data[key])) {
      (data as unknown as Record<string, unknown>)[key] = [];
    }
  }
  return data;
}

export async function fetchSnapshot(): Promise<Snapshot> {
  const response = await fetch("/api/snapshot");
  if (!response.ok) throw new Error("snapshot failed");
  return normalize(await response.json());
}

export async function dispatchAction(payload: ActionRequest): Promise<Snapshot> {
  const response = await fetch("/api/action", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const data = (await response.json()) as Snapshot & { error?: string };
  if (!response.ok || data.ok === false) {
    throw new Error(data.error || "action failed");
  }
  return normalize(data);
}
