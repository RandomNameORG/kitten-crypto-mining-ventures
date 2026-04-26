import { ActionBar, ActionButton } from "../components/ActionButton";
import type { Snapshot } from "../types";
import { gpuIconSrc } from "../util";

interface Props {
  snapshot: Snapshot;
  dispatch: (id: string) => void;
}

export function StorePanel({ snapshot, dispatch }: Props) {
  return (
    <>
      <div className="list">
        {snapshot.gpu_defs.map((def) => {
          const canBuy = snapshot.state.btc >= def.price;
          return (
            <article key={def.id} className="row item-row">
              <img className="item-icon gpu-icon" src={gpuIconSrc(def.id)} alt="" loading="lazy" />
              <div className="item-content">
                <div className="row-head">
                  <span className="row-title">{def.name}</span>
                  <span className="tag">{def.tier}</span>
                </div>
                <div className="copy">{def.flavor}</div>
                <div className="facts">
                  <span className="fact price">{def.price_fmt}</span>
                  <span className="fact">eff {def.efficiency.toFixed(4)}</span>
                  <span className="fact">heat {def.heat_output.toFixed(2)}</span>
                </div>
                <ActionBar>
                  <ActionButton
                    label={canBuy ? "购买" : "余额不足"}
                    icon="买"
                    intent="primary"
                    disabled={!canBuy}
                    onClick={() => dispatch(def.id)}
                  />
                </ActionBar>
              </div>
            </article>
          );
        })}
      </div>
    </>
  );
}
