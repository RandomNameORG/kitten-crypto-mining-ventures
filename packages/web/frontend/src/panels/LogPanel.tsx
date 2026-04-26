import type { Snapshot } from "../types";

interface Props {
  snapshot: Snapshot;
}

export function LogPanel({ snapshot }: Props) {
  return (
    <>
      <h2>日志</h2>
      <div>
        {snapshot.log.map((entry, i) => (
          <div key={`${entry.time}-${i}`} className="logline">
            [{entry.category}] {entry.text}
          </div>
        ))}
      </div>
    </>
  );
}
