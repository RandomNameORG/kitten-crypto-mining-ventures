interface Props {
  items: Array<readonly [label: string, value: string]>;
}

export function PanelSummary({ items }: Props) {
  return (
    <div className="panel-summary">
      {items.map(([label, value]) => (
        <div key={label} className="summary-chip">
          <span>{label}</span>
          <strong>{value}</strong>
        </div>
      ))}
    </div>
  );
}
