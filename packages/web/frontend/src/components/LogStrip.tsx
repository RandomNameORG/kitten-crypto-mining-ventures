import type { LogEntry } from "../types";

interface Props {
  log: LogEntry[];
}

export function LogStrip({ log }: Props) {
  const recent = log.slice(-6).reverse();
  if (recent.length === 0) {
    return (
      <div className="log-strip">
        <div className="log-strip-line">
          <span className="log-cat">log</span>
          <span className="log-text">…</span>
        </div>
      </div>
    );
  }
  return (
    <div className="log-strip">
      {recent.map((entry, i) => (
        <div key={`${entry.time}-${i}`} className="log-strip-line">
          <span className={`log-cat ${entry.category}`}>{entry.category}</span>
          <span className="log-text">{entry.text}</span>
        </div>
      ))}
    </div>
  );
}
