import { useMemo, useState } from "react";
import { EventBanner } from "./components/EventBanner";
import { GameStage } from "./components/GameStage";
import { Hud } from "./components/Hud";
import { LogStrip } from "./components/LogStrip";
import { StageFoot } from "./components/StageFoot";
import { Tabs } from "./components/Tabs";
import { useSnapshot } from "./hooks/useSnapshot";
import { DefensePanel } from "./panels/DefensePanel";
import { GPUsPanel } from "./panels/GPUsPanel";
import { LogPanel } from "./panels/LogPanel";
import { MercsPanel } from "./panels/MercsPanel";
import { RoomsPanel } from "./panels/RoomsPanel";
import { SkillsPanel } from "./panels/SkillsPanel";
import { StatsPanel } from "./panels/StatsPanel";
import { StorePanel } from "./panels/StorePanel";
import type { Snapshot, TabId } from "./types";

export function App() {
  const { snapshot, message, dispatch } = useSnapshot();
  const [tab, setTab] = useState<TabId>("store");

  const room = useMemo(
    () => snapshot?.rooms.find((r) => r.id === snapshot.state.current_room) ?? null,
    [snapshot],
  );
  const roomName = room?.name ?? "loading";
  const roomFlavor = room?.flavor ?? "loading";
  const status = !snapshot
    ? { tone: "warn", label: "CONNECTING" }
    : snapshot.state.paused
      ? { tone: "warn", label: "PAUSED" }
      : snapshot.state.mining_paused
        ? { tone: "warn", label: "REBOOTING" }
        : { tone: "live", label: "MINING" };
  const toast = snapshot?.last_event
    ? `${snapshot.last_event.name}: ${snapshot.last_event.text}`
    : message;

  return (
    <main className="app">
      <header className="topbar">
        <div className="brand">
          <div className="catmark">M</div>
          <div>
            <h1>矿业大亨喵</h1>
            <div className="subtitle">
              <span className={`status-pill ${status.tone}`}>
                <span className={`status-dot ${status.tone === "warn" ? "warn" : ""}`} />
                {status.label}
              </span>
              <span>· {snapshot?.state.kitten_name ?? "—"}</span>
            </div>
          </div>
        </div>
        {snapshot && <Hud state={snapshot.state} room={room} />}
      </header>

      <section className="layout">
        <section className="stage-shell">
          <div className="stage-head">
            <div className="room-title">
              <strong>{roomName}</strong>
              <span>{roomFlavor}</span>
            </div>
            <div className="stage-actions">
              <button type="button" onClick={() => dispatch({ action: "toggle_pause" })}>
                {snapshot?.state.paused ? "继续" : "暂停"}
              </button>
              <button type="button" onClick={() => dispatch({ action: "vent" })}>
                排热
              </button>
              <button type="button" onClick={() => dispatch({ action: "reset" })}>
                重开
              </button>
            </div>
          </div>
          <div className="canvas-wrap">
            <GameStage snapshot={snapshot} />
            <div className="toast">{toast}</div>
          </div>
          <StageFoot room={room} />
          <EventBanner event={snapshot?.last_event} />
          <LogStrip log={snapshot?.log ?? []} />
        </section>

        <aside className="side">
          <Tabs active={tab} onSelect={setTab} />
          {snapshot && (
            <PanelHeader tab={tab} snapshot={snapshot} />
          )}
          <section className="panel">
            {!snapshot ? (
              <p>{message}</p>
            ) : tab === "store" ? (
              <StorePanel
                snapshot={snapshot}
                dispatch={(id) => dispatch({ action: "buy_gpu", id })}
              />
            ) : tab === "rooms" ? (
              <RoomsPanel snapshot={snapshot} dispatch={dispatch} />
            ) : tab === "gpus" ? (
              <GPUsPanel snapshot={snapshot} dispatch={dispatch} />
            ) : tab === "defense" ? (
              <DefensePanel snapshot={snapshot} dispatch={dispatch} />
            ) : tab === "skills" ? (
              <SkillsPanel snapshot={snapshot} dispatch={dispatch} />
            ) : tab === "mercs" ? (
              <MercsPanel snapshot={snapshot} dispatch={dispatch} />
            ) : tab === "log" ? (
              <LogPanel snapshot={snapshot} />
            ) : (
              <StatsPanel snapshot={snapshot} />
            )}
          </section>
        </aside>
      </section>
    </main>
  );
}

function PanelHeader({ tab, snapshot }: { tab: TabId; snapshot: Snapshot }) {
  const room = snapshot.rooms.find((r) => r.id === snapshot.state.current_room);
  const ownedGpus = snapshot.gpus.filter((g) => g.room === snapshot.state.current_room);
  const broken = ownedGpus.filter((g) => g.status === "broken").length;
  const learnedSkills = snapshot.skills.filter((s) => s.unlocked).length;

  const heading: Record<TabId, { title: string; meta: string }> = {
    store: { title: "显卡商店", meta: `${snapshot.gpu_defs.length} 款 · 余额 ${snapshot.state.btc_fmt}` },
    rooms: { title: "房间矩阵", meta: `${snapshot.rooms.filter((r) => r.unlocked).length}/${snapshot.rooms.length} 已解锁` },
    gpus: { title: "当前房间机架", meta: room ? `${room.gpu_count}/${room.slots} 槽位${broken ? ` · ${broken} 损坏` : ""}` : "" },
    defense: { title: "防御与维护", meta: room ? `${room.name} · ${room.heat.toFixed(0)}°/${room.max_heat.toFixed(0)}°` : "" },
    skills: { title: "技能树", meta: `${learnedSkills}/${snapshot.skills.length} 已学 · TP ${snapshot.state.tech_point}` },
    mercs: { title: "雇佣猫", meta: `${snapshot.mercs.length} 人在职 · ${snapshot.merc_defs.length} 类可雇` },
    log: { title: "日志", meta: `${snapshot.log.length} 条记录` },
    stats: { title: "运营状态", meta: `累计收益 ${snapshot.state.lifetime_earned_fmt}` },
  };
  const h = heading[tab];

  return (
    <div className="panel-header">
      <span className="panel-title">{h.title}</span>
      <span className="panel-meta">{h.meta}</span>
    </div>
  );
}
