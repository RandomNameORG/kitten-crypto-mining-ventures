export function App() {
  return (
    <main className="app">
      <header className="topbar">
        <div className="brand">
          <div className="catmark">M</div>
          <div>
            <h1>矿业大亨喵</h1>
            <div className="subtitle">2D Pixel Tycoon · React shell</div>
          </div>
        </div>
        <section className="hud">
          <div className="metric">
            <span className="metric-label">phase</span>
            <span className="metric-value">1 — scaffold</span>
          </div>
        </section>
      </header>
      <section className="layout">
        <section className="stage-shell">
          <div className="stage-head">
            <div className="room-title">
              <strong>React shell ready</strong>
              <span>side panel + canvas migration pending</span>
            </div>
          </div>
          <div className="canvas-wrap">
            <div className="placeholder">
              GameStage placeholder · phase 3 ports the canvas
            </div>
          </div>
        </section>
        <aside className="side">
          <nav className="tabs">
            <span className="tab active">store</span>
            <span className="tab">rooms</span>
            <span className="tab">gpus</span>
          </nav>
          <section className="panel">
            <p className="muted">
              panel components arrive in phase 2.
            </p>
          </section>
        </aside>
      </section>
    </main>
  );
}
