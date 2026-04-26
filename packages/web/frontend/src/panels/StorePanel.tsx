import { useMemo, useRef, useState } from "react";
import { ActionBar, ActionButton } from "../components/ActionButton";
import type { Snapshot } from "../types";
import { gpuIconSrc } from "../util";

interface Props {
  snapshot: Snapshot;
  dispatch: (id: string) => void;
}

export function StorePanel({ snapshot, dispatch }: Props) {
  const [pending, setPending] = useState<string | null>(null);
  const lastClickAt = useRef(0);

  const ownedByDef = useMemo(() => {
    const map = new Map<string, { running: number; shipping: number; broken: number }>();
    for (const g of snapshot.gpus) {
      const k = g.def_id;
      if (!k) continue;
      const cur = map.get(k) ?? { running: 0, shipping: 0, broken: 0 };
      if (g.status === "running") cur.running += 1;
      else if (g.status === "shipping") cur.shipping += 1;
      else if (g.status === "broken") cur.broken += 1;
      map.set(k, cur);
    }
    return map;
  }, [snapshot.gpus]);

  const handleBuy = (id: string) => {
    const now = Date.now();
    if (now - lastClickAt.current < 350) return;
    lastClickAt.current = now;
    setPending(id);
    dispatch(id);
    window.setTimeout(() => setPending(null), 600);
  };

  return (
    <div className="list">
      {snapshot.gpu_defs.map((def) => {
        const canBuy = snapshot.state.btc >= def.price;
        const owned = ownedByDef.get(def.id);
        const inFlight = pending === def.id;
        const label = inFlight ? "下单中…" : canBuy ? "购买" : "余额不足";
        return (
          <article key={def.id} className={`row item-row tier-${def.tier}`}>
            <img className="item-icon gpu-icon" src={gpuIconSrc(def.id)} alt="" loading="lazy" />
            <div className="item-content">
              <div className="row-head">
                <span className="row-title">{def.name}</span>
                <span className={`tag tier tier-${def.tier}`}>{def.tier}</span>
              </div>
              <div className="copy">{def.flavor}</div>
              <div className="facts">
                <span className="fact price">{def.price_fmt}</span>
                <span className="fact">eff {def.efficiency.toFixed(4)}</span>
                <span className="fact">heat {def.heat_output.toFixed(2)}</span>
                {owned && (owned.running + owned.shipping + owned.broken) > 0 && (
                  <span className="fact owned">
                    已有 {owned.running}
                    {owned.shipping > 0 && ` · 运输 ${owned.shipping}`}
                    {owned.broken > 0 && ` · 坏 ${owned.broken}`}
                  </span>
                )}
              </div>
              <ActionBar>
                <ActionButton
                  label={label}
                  icon="买"
                  intent="primary"
                  disabled={!canBuy || inFlight}
                  onClick={() => handleBuy(def.id)}
                />
              </ActionBar>
            </div>
          </article>
        );
      })}
    </div>
  );
}
