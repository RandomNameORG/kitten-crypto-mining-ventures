import { useCallback, useEffect, useRef, useState } from "react";
import { dispatchAction, fetchSnapshot } from "../api";
import type { ActionRequest, Snapshot } from "../types";

export interface SnapshotStore {
  snapshot: Snapshot | null;
  message: string;
  dispatch: (payload: ActionRequest) => Promise<void>;
}

export function useSnapshot(pollMs = 1000): SnapshotStore {
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  const [message, setMessage] = useState<string>("connecting");
  const inflight = useRef(false);

  const refresh = useCallback(async () => {
    if (inflight.current) return;
    inflight.current = true;
    try {
      const data = await fetchSnapshot();
      setSnapshot(data);
      setMessage((prev) => (prev === "connecting" ? "ready" : prev));
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err));
    } finally {
      inflight.current = false;
    }
  }, []);

  useEffect(() => {
    refresh();
    const id = window.setInterval(refresh, pollMs);
    return () => window.clearInterval(id);
  }, [refresh, pollMs]);

  const dispatch = useCallback(async (payload: ActionRequest) => {
    try {
      const data = await dispatchAction(payload);
      setSnapshot(data);
      setMessage("done");
    } catch (err) {
      setMessage(err instanceof Error ? err.message : String(err));
    }
  }, []);

  return { snapshot, message, dispatch };
}
