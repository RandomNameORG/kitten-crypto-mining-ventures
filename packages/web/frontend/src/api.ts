import type { ActionRequest, Snapshot } from "./types";

export async function fetchSnapshot(): Promise<Snapshot> {
  const response = await fetch("/api/snapshot");
  if (!response.ok) throw new Error("snapshot failed");
  return response.json();
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
  return data;
}
