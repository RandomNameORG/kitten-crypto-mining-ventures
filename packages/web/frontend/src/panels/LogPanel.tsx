import type { Snapshot } from "../types";

interface Props {
  snapshot: Snapshot;
}

export function LogPanel({ snapshot }: Props) {
  return (
    <div className="flex flex-col gap-0">
      {snapshot.log.slice().reverse().map((entry, i) => (
        <div key={`${entry.time}-${i}`} className="logline">
          <span className="text-muted">[{entry.category}]</span> {entry.text}
        </div>
      ))}
    </div>
  );
}
