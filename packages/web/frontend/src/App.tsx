import { useMemo, useState } from "react";
import { GameStage } from "./components/GameStage";
import { Hud } from "./components/Hud";
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
import type { TabId } from "./types";

export function App() {
  const { snapshot, message, dispatch } = useSnapshot();
  const [tab, setTab] = useState<TabId>("store");

  const room = useMemo(
    () => snapshot?.rooms.find((r) => r.id === snapshot.state.current_room) ?? null,
    [snapshot],
  );
  const roomName = room?.name ?? "loading";
  const roomFlavor = room?.flavor ?? "loading";
  const subtitle = snapshot
    ? `${snapshot.state.kitten_name} · ${snapshot.state.paused ? "paused" : "mining"}`
    : "connecting";
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
            <div className="subtitle">{subtitle}</div>
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
        </section>

        <aside className="side">
          <Tabs active={tab} onSelect={setTab} />
          <section className="panel">
            {!snapshot ? (
              <p>{message}</p>
            ) : tab === "store" ? (
              <StorePanel
                snapshot={snapshot}
                currentRoomName={roomName}
                dispatch={(id) => dispatch({ action: "buy_gpu", id })}
              />
            ) : tab === "rooms" ? (
              <RoomsPanel snapshot={snapshot} currentRoomName={roomName} dispatch={dispatch} />
            ) : tab === "gpus" ? (
              <GPUsPanel snapshot={snapshot} dispatch={dispatch} />
            ) : tab === "defense" ? (
              <DefensePanel snapshot={snapshot} dispatch={dispatch} />
            ) : tab === "skills" ? (
              <SkillsPanel snapshot={snapshot} dispatch={dispatch} />
            ) : tab === "mercs" ? (
              <MercsPanel snapshot={snapshot} currentRoomName={roomName} dispatch={dispatch} />
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
