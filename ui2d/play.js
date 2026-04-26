const canvas = document.querySelector("#stage");
const ctx = canvas.getContext("2d");
const stageShell = document.querySelector(".stage-shell");
const side = document.querySelector(".side");
const hud = document.querySelector("#hud");
const stageFoot = document.querySelector("#stageFoot");
const panel = document.querySelector("#panel");
const tabsEl = document.querySelector("#tabs");
const toast = document.querySelector("#toast");
const roomName = document.querySelector("#roomName");
const roomFlavor = document.querySelector("#roomFlavor");
const subtitle = document.querySelector("#subtitle");
const pauseButton = document.querySelector("#pauseButton");
const ventButton = document.querySelector("#ventButton");
const resetButton = document.querySelector("#resetButton");

const STAGE = { width: 512, height: 288 };
const DIRECTIONS = ["down", "left", "right", "up"];
const FRAME_COUNT = 4;
const ARRIVAL_DISTANCE = 6;
const CHARACTER_ROOT = "/assets/2d/spritesheet/characters";
const tabs = [
  ["store", "商店", "S"],
  ["rooms", "房间", "R"],
  ["gpus", "显卡", "G"],
  ["defense", "防御", "D"],
  ["skills", "技能", "T"],
  ["mercs", "雇佣", "H"],
  ["log", "日志", "L"],
  ["stats", "状态", "I"],
];

let model = null;
let sceneMap = {
  actor: {
    spawn: { x: 130, y: 220 },
    walk_polygon: [[24, 184], [488, 184], [504, 254], [10, 254]],
  },
  room_walk_polygons: {},
  cat_points: {
    default: {
      door: [70, 238],
      center: [246, 236],
      rig: [352, 228],
      fan: [304, 232],
      rest: [136, 238],
    },
    rooms: {},
  },
  rig_slots: {
    origin_x: 286,
    origin_y: 112,
    columns: 6,
    column_gap: 34,
    row_gap: 35,
    width: 28,
    height: 22,
  },
};
let activeTab = "store";
let lastMessage = "ready";
let lastEventKey = "";
let previousTime = performance.now();
const keys = new Set();
const pointer = { active: false, x: 248, y: 212 };
const actor = {
  x: 130,
  y: 220,
  targetX: 130,
  targetY: 220,
  direction: "right",
  frame: 0,
  frameClock: 0,
  moving: false,
  speed: 116,
  scale: 0.72,
  targetBodyHeight: 82,
  state: "patrol",
  stateClock: 0,
  decisionClock: 0,
  reactionClock: 0,
  emote: "",
  emoteClock: 0,
  bobClock: 0,
  coolingIndex: 0,
};
const images = {
  rooms: new Map(),
  sprites: new Map(),
};

ctx.imageSmoothingEnabled = false;

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function loadImage(src) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error(`asset ${src}`));
    img.src = src;
  });
}

async function ensureRoomImage(room) {
  if (!room || images.rooms.has(room.id)) return;
  images.rooms.set(room.id, await loadImage(room.background));
}

async function ensureSprite(mode, direction, frame) {
  const key = `${mode}:${direction}:${frame}`;
  if (images.sprites.has(key)) return;
  const src = `${CHARACTER_ROOT}/player/${mode}/${direction}-${frame}.png`;
  const image = await loadImage(src);
  images.sprites.set(key, extractSpriteBody(image));
}

function extractSpriteBody(image) {
  const scratch = document.createElement("canvas");
  scratch.width = image.width;
  scratch.height = image.height;
  const scratchCtx = scratch.getContext("2d", { willReadFrequently: true });
  scratchCtx.drawImage(image, 0, 0);
  const frame = scratchCtx.getImageData(0, 0, image.width, image.height);
  const pixels = frame.data;
  const total = image.width * image.height;
  const visited = new Uint8Array(total);
  let best = null;

  for (let start = 0; start < total; start += 1) {
    if (visited[start] || pixels[start * 4 + 3] < 8) continue;
    const stack = [start];
    const component = [];
    visited[start] = 1;
    let left = image.width;
    let right = 0;
    let top = image.height;
    let bottom = 0;

    while (stack.length) {
      const index = stack.pop();
      component.push(index);
      const x = index % image.width;
      const y = Math.floor(index / image.width);
      if (x < left) left = x;
      if (x + 1 > right) right = x + 1;
      if (y < top) top = y;
      if (y + 1 > bottom) bottom = y + 1;

      for (let ny = y - 1; ny <= y + 1; ny += 1) {
        for (let nx = x - 1; nx <= x + 1; nx += 1) {
          if (nx < 0 || ny < 0 || nx >= image.width || ny >= image.height) continue;
          const next = ny * image.width + nx;
          if (visited[next] || pixels[next * 4 + 3] < 8) continue;
          visited[next] = 1;
          stack.push(next);
        }
      }
    }

    if (!best || component.length > best.component.length) {
      best = { component, bounds: { x: left, y: top, width: right - left, height: bottom - top } };
    }
  }

  if (!best) {
    return { image, bounds: { x: 0, y: 0, width: image.width, height: image.height } };
  }

  const keep = new Uint8Array(total);
  for (const index of best.component) keep[index] = 1;
  for (let index = 0; index < total; index += 1) {
    if (!keep[index]) pixels[index * 4 + 3] = 0;
  }
  scratchCtx.putImageData(frame, 0, 0);
  return { image: scratch, bounds: best.bounds };
}

async function preloadSprites() {
  const tasks = [];
  for (const mode of ["idle", "walk"]) {
    for (const direction of DIRECTIONS) {
      for (let frame = 1; frame <= FRAME_COUNT; frame += 1) {
        tasks.push(ensureSprite(mode, direction, frame));
      }
    }
  }
  await Promise.all(tasks);
}

async function loadSceneMap() {
  const response = await fetch("/ui2d/scene-map.json");
  if (!response.ok) throw new Error("scene map failed");
  sceneMap = await response.json();
  actor.x = sceneMap.actor.spawn.x;
  actor.y = sceneMap.actor.spawn.y;
  actor.targetX = actor.x;
  actor.targetY = actor.y;
  pointer.x = actor.x;
  pointer.y = actor.y;
}

function currentRoom() {
  return model?.rooms.find((room) => room.id === model.state.current_room);
}

function roomById(id) {
  return model?.rooms.find((room) => room.id === id);
}

function currentWalkPolygon() {
  const room = currentRoom();
  return sceneMap.room_walk_polygons?.[room?.id] || sceneMap.actor.walk_polygon;
}

function catPoint(name) {
  const room = currentRoom();
  const roomPoints = sceneMap.cat_points?.rooms?.[room?.id] || {};
  const point = roomPoints[name] || sceneMap.cat_points?.default?.[name] || sceneMap.actor.spawn;
  return Array.isArray(point) ? { x: point[0], y: point[1] } : { x: point.x, y: point.y };
}

function normalizeModel(data) {
  if (!data) return data;
  for (const key of ["rooms", "gpus", "gpu_defs", "skills", "mercs", "merc_defs", "log"]) {
    if (!Array.isArray(data[key])) data[key] = [];
  }
  return data;
}

function syncSideHeight() {
  const height = Math.round(stageShell.getBoundingClientRect().height);
  if (height > 0) side.style.setProperty("--stage-height", `${height}px`);
}

function pct(value) {
  return `${Math.max(0, Math.min(100, value * 100)).toFixed(0)}%`;
}

async function fetchSnapshot() {
  const response = await fetch("/api/snapshot");
  if (!response.ok) throw new Error("snapshot failed");
  model = normalizeModel(await response.json());
  await ensureRoomImage(currentRoom());
  render();
}

async function action(payload) {
  const response = await fetch("/api/action", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const data = await response.json();
  if (!response.ok || data.ok === false) {
    lastMessage = data.error || "action failed";
    toast.textContent = lastMessage;
    return;
  }
  model = normalizeModel(data);
  await ensureRoomImage(currentRoom());
  lastMessage = "done";
  render();
}

function initTabs() {
  tabsEl.innerHTML = tabs.map(([id, label, icon]) => (
    `<button type="button" class="${id === activeTab ? "active" : ""}" data-tab="${id}" title="${label}">
      <span class="tab-icon" aria-hidden="true">${icon}</span>
      <span class="tab-label">${label}</span>
    </button>`
  )).join("");
  tabsEl.querySelectorAll("button").forEach((button) => {
    button.addEventListener("click", () => {
      activeTab = button.dataset.tab;
      initTabs();
      renderPanel();
    });
  });
}

function render() {
  if (!model) return;
  const room = currentRoom();
  subtitle.textContent = `${model.state.kitten_name} · ${model.state.paused ? "paused" : "mining"}`;
  roomName.textContent = room?.name || "unknown";
  roomFlavor.textContent = room?.flavor || "";
  pauseButton.textContent = model.state.paused ? "继续" : "暂停";
  toast.textContent = model.last_event ? `${model.last_event.name}: ${model.last_event.text}` : lastMessage;
  renderHud();
  renderStageFoot(room);
  renderPanel();
  syncSideHeight();
}

function renderHud() {
  const state = model.state;
  const room = currentRoom();
  const heat = room ? `${room.heat.toFixed(0)}° / ${room.max_heat.toFixed(0)}°` : "-";
  const trend = state.market_trend > 0 ? "↑" : state.market_trend < 0 ? "↓" : "→";
  hud.innerHTML = [
    metric("BTC", state.btc_fmt),
    metric("净收益", state.room_net_fmt),
    metric("产出", state.room_earn_fmt),
    metric("电费", state.room_bill_fmt),
    metric("温度", heat, room && room.heat_pct > 0.78 ? "hot" : ""),
    metric("市场", `${state.market_price.toFixed(2)} ${trend}`),
  ].join("");
}

function metric(label, value, cls = "") {
  return `<div class="metric"><span>${label}</span><strong class="${cls}">${escapeHtml(value)}</strong></div>`;
}

function renderStageFoot(room) {
  if (!room) return;
  const slots = room.slots ? room.gpu_count / room.slots : 0;
  const defense = room.defense || {};
  const shield = ((defense.lock || 0) + (defense.cctv || 0) + (defense.armor || 0)) / 24;
  const cooling = ((defense.cooling || 0) + (defense.wiring || 0)) / 16;
  stageFoot.innerHTML = [
    bar("槽位", `${room.gpu_count}/${room.slots}`, slots),
    bar("温度", `${room.heat.toFixed(0)}°`, room.heat_pct, "heat"),
    bar("安防", `${Math.round(shield * 100)}%`, shield),
    bar("维护", `${Math.round(cooling * 100)}%`, cooling),
  ].join("");
}

function bar(label, value, amount, cls = "") {
  return `<div class="bar">
    <div class="bar-label"><span>${label}</span><span>${value}</span></div>
    <div class="bar-track"><div class="bar-fill ${cls}" style="width:${pct(amount)}"></div></div>
  </div>`;
}

function actionButton({ action, label, id = "", dim = "", instance = "", disabled = false, intent = "default", icon = "" }) {
  const attrs = [
    `type="button"`,
    `class="action-btn ${intent}"`,
    `data-action="${action}"`,
  ];
  if (id) attrs.push(`data-id="${escapeHtml(id)}"`);
  if (dim) attrs.push(`data-dim="${escapeHtml(dim)}"`);
  if (instance) attrs.push(`data-instance="${instance}"`);
  if (disabled) attrs.push("disabled");
  return `<button ${attrs.join(" ")}>
    <span class="action-icon" aria-hidden="true">${escapeHtml(icon || label.slice(0, 1))}</span>
    <span>${escapeHtml(label)}</span>
  </button>`;
}

function actionBar(buttons) {
  return `<div class="actions">${buttons.join("")}</div>`;
}

function panelSummary(items) {
  return `<div class="panel-summary">${items.map(([label, value, cls = ""]) => (
    `<div class="summary-chip ${cls}"><span>${escapeHtml(label)}</span><strong>${escapeHtml(value)}</strong></div>`
  )).join("")}</div>`;
}

function statusText(ok, yes, no) {
  return ok ? yes : no;
}

function renderPanel() {
  if (!model) return;
  const renderers = {
    store: renderStore,
    rooms: renderRooms,
    gpus: renderGPUs,
    defense: renderDefense,
    skills: renderSkills,
    mercs: renderMercs,
    log: renderLog,
    stats: renderStats,
  };
  panel.innerHTML = renderers[activeTab]();
  panel.querySelectorAll("[data-action]").forEach((button) => {
    button.addEventListener("click", () => {
      const payload = { action: button.dataset.action };
      if (button.dataset.id) payload.id = button.dataset.id;
      if (button.dataset.dim) payload.dim = button.dataset.dim;
      if (button.dataset.instance) payload.instance_id = Number(button.dataset.instance);
      action(payload);
    });
  });
}

function renderStore() {
  const room = currentRoom();
  return `<h2>显卡商店</h2>${panelSummary([
    ["余额", model.state.btc_fmt],
    ["当前房间", room?.name || "-"],
    ["槽位", room ? `${room.gpu_count}/${room.slots}` : "-"],
  ])}<div class="list">${model.gpu_defs.map((def) => {
    const canBuy = model.state.btc >= def.price;
    return `<article class="row">
      <div class="row-head"><span class="row-title">${escapeHtml(def.name)}</span><span class="tag">${escapeHtml(def.tier)}</span></div>
      <div class="copy">${escapeHtml(def.flavor)}</div>
      <div class="facts">
        <span class="fact price">${def.price_fmt}</span>
        <span class="fact">eff ${def.efficiency.toFixed(4)}</span>
        <span class="fact">heat ${def.heat_output.toFixed(2)}</span>
      </div>
      ${actionBar([
        actionButton({ action: "buy_gpu", id: def.id, label: canBuy ? "购买" : "余额不足", disabled: !canBuy, intent: "primary", icon: "买" }),
      ])}
    </article>`;
  }).join("")}</div>`;
}

function renderRooms() {
  return `<h2>房间</h2>${panelSummary([
    ["当前", currentRoom()?.name || "-"],
    ["余额", model.state.btc_fmt],
  ])}<div class="list">${model.rooms.map((room) => {
    const actionButton = room.unlocked
      ? actionButtonMarkup("switch_room", room.id, room.current ? "已在此处" : "进入", room.current, "primary", "入")
      : actionButtonMarkup("unlock_room", room.id, model.state.btc >= room.unlock_cost ? "解锁" : "余额不足", model.state.btc < room.unlock_cost, "primary", "解");
    return `<article class="row">
      <div class="row-head"><span class="row-title">${escapeHtml(room.name)}</span><span class="tag">${room.unlocked ? `${room.gpu_count}/${room.slots}` : room.unlock_cost_fmt}</span></div>
      <div class="copy">${escapeHtml(room.flavor)}</div>
      <div class="facts">
        <span class="fact">net ${room.net_fmt}</span>
        <span class="fact">heat ${room.heat ? room.heat.toFixed(0) : 0}°</span>
        <span class="fact">tick ${room.heat_tick_in || 0}s</span>
      </div>
      ${actionBar([actionButton])}
    </article>`;
  }).join("")}</div>`;
}

function actionButtonMarkup(action, id, label, disabled, intent, icon) {
  return actionButton({ action, id, label, disabled, intent, icon });
}

function renderGPUs() {
  const gpus = model.gpus.filter((gpu) => gpu.room === model.state.current_room);
  const room = currentRoom();
  if (!gpus.length) return `<h2>当前房间显卡</h2>${panelSummary([["槽位", room ? `0/${room.slots}` : "0"], ["提示", "去商店购买"]])}<div class="empty">空槽位</div>`;
  return `<h2>当前房间显卡</h2>${panelSummary([
    ["槽位", room ? `${room.gpu_count}/${room.slots}` : `${gpus.length}`],
    ["损坏", `${gpus.filter((gpu) => gpu.status === "broken").length}`],
  ])}<div class="list">${gpus.map((gpu) => `<article class="row">
    <div class="row-head"><span class="row-title">#${gpu.instance_id} ${escapeHtml(gpu.name)}</span><span class="tag">${escapeHtml(gpu.status)}</span></div>
    <div class="facts">
      <span class="fact">L${gpu.upgrade}</span>
      <span class="fact">OC ${gpu.oc_level}</span>
      <span class="fact">${gpu.earn_fmt}</span>
      <span class="fact">${gpu.hours_left.toFixed(1)}h</span>
    </div>
    ${actionBar([
      actionButton({ action: "upgrade_gpu", instance: gpu.instance_id, label: "升级", intent: "primary", icon: "升" }),
      actionButton({ action: "cycle_oc", instance: gpu.instance_id, label: "超频", intent: "accent", icon: "频" }),
      actionButton({ action: "repair_gpu", instance: gpu.instance_id, label: gpu.repairable ? "维修" : "正常", disabled: !gpu.repairable, intent: "warn", icon: "修" }),
      actionButton({ action: "scrap_gpu", instance: gpu.instance_id, label: "拆解", intent: "danger", icon: "拆" }),
    ])}
  </article>`).join("")}</div>`;
}

function renderDefense() {
  const room = currentRoom();
  const d = room.defense || {};
  const dims = [
    ["lock", "门锁", d.lock || 0],
    ["cctv", "监控", d.cctv || 0],
    ["wiring", "布线", d.wiring || 0],
    ["cooling", "散热", d.cooling || 0],
    ["armor", "装甲", d.armor || 0],
  ];
  return `<h2>防御与维护</h2>${panelSummary([
    ["温度", room ? `${room.heat.toFixed(0)}°/${room.max_heat.toFixed(0)}°` : "-"],
    ["余额", model.state.btc_fmt],
  ])}<div class="list">${dims.map(([id, label, level]) => `<article class="row">
    <div class="row-head"><span class="row-title">${label}</span><span class="tag">L${level}</span></div>
    <div class="facts"><span class="fact">cost ${(level + 1) * 250}</span><span class="fact">max 8</span></div>
    ${actionBar([
      actionButton({ action: "upgrade_defense", dim: id, label: level >= 8 ? "已满级" : "升级", disabled: level >= 8, intent: "primary", icon: "升" }),
    ])}
  </article>`).join("")}</div>`;
}

function renderSkills() {
  const visible = model.skills.slice(0, 18);
  return `<h2>技能</h2>${panelSummary([
    ["TP", `${model.state.tech_point}`],
    ["碎片", `${model.state.research_frags}`],
  ])}<div class="list">${visible.map((skill) => {
    const prereqOk = !skill.prereq || model.skills.find((item) => item.id === skill.prereq)?.unlocked;
    const canBuy = !skill.unlocked && prereqOk && model.state.tech_point >= skill.cost;
    return `<article class="row">
      <div class="row-head"><span class="row-title">${escapeHtml(skill.name)}</span><span class="tag">${skill.cost} TP</span></div>
      <div class="copy">${escapeHtml(skill.desc)}</div>
      <div class="facts"><span class="fact">${escapeHtml(skill.lane)}</span><span class="fact">${statusText(skill.unlocked, "已学会", prereqOk ? "可研究" : "前置未解")}</span></div>
      ${actionBar([
        actionButton({ action: "unlock_skill", id: skill.id, label: skill.unlocked ? "已研究" : "研究", disabled: !canBuy, intent: "primary", icon: "研" }),
      ])}
    </article>`;
  }).join("")}</div>`;
}

function renderMercs() {
  const owned = model.mercs.map((merc) => `<article class="row">
    <div class="row-head"><span class="row-title">#${merc.instance_id} ${escapeHtml(merc.name)}</span><span class="tag">${merc.loyalty}</span></div>
    <div class="facts"><span class="fact">${escapeHtml(roomById(merc.room_id)?.name || merc.room_id)}</span></div>
    ${actionBar([
      actionButton({ action: "bribe_merc", instance: merc.instance_id, label: "打赏", intent: "accent", icon: "赏" }),
      actionButton({ action: "fire_merc", instance: merc.instance_id, label: "解雇", intent: "danger", icon: "离" }),
    ])}
  </article>`).join("");
  const hire = model.merc_defs.map((merc) => `<article class="row">
    <div class="row-head"><span class="row-title">${escapeHtml(merc.name)}</span><span class="tag">${merc.hire_cost_fmt}</span></div>
    <div class="copy">${escapeHtml(merc.flavor)}</div>
    <div class="facts"><span class="fact">${escapeHtml(merc.specialty)}</span><span class="fact">wage ${merc.wage_fmt}</span></div>
    ${actionBar([
      actionButton({ action: "hire_merc", id: merc.id, label: "雇佣", intent: "primary", icon: "雇" }),
    ])}
  </article>`).join("");
  return `<h2>雇佣猫</h2>${panelSummary([["已雇佣", `${model.mercs.length}`], ["当前房间", currentRoom()?.name || "-"]])}<div class="list">${owned || "<div class=\"empty\">暂无雇佣</div>"}${hire}</div>`;
}

function renderLog() {
  return `<h2>日志</h2><div>${model.log.map((entry) => `<div class="logline">[${escapeHtml(entry.category)}] ${escapeHtml(entry.text)}</div>`).join("")}</div>`;
}

function renderStats() {
  const room = currentRoom();
  const allGPUs = model.gpus.length;
  const broken = model.gpus.filter((gpu) => gpu.status === "broken").length;
  return `<h2>状态</h2><div class="list">
    <article class="row"><div class="row-head"><span class="row-title">资产</span><span class="tag">${allGPUs}</span></div>
      <div class="facts"><span class="fact">broken ${broken}</span><span class="fact">TP ${model.state.tech_point}</span><span class="fact">frags ${model.state.research_frags}</span></div></article>
    <article class="row"><div class="row-head"><span class="row-title">${escapeHtml(room.name)}</span><span class="tag">${room.net_fmt}</span></div>
      <div class="facts"><span class="fact">earn ${room.earn_fmt}</span><span class="fact">bill ${room.bill_fmt}</span><span class="fact">heat ${room.heat_delta.toFixed(1)}</span></div></article>
    <article class="row"><div class="row-head"><span class="row-title">声望</span><span class="tag">${model.state.reputation}</span></div>
      <div class="facts"><span class="fact">karma ${model.state.karma}</span><span class="fact">life ${model.state.lifetime_earned_fmt}</span></div></article>
  </div>`;
}

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value));
}

function polygonBounds(points) {
  return points.reduce((bounds, [x, y]) => ({
    minX: Math.min(bounds.minX, x),
    maxX: Math.max(bounds.maxX, x),
    minY: Math.min(bounds.minY, y),
    maxY: Math.max(bounds.maxY, y),
  }), { minX: Infinity, maxX: -Infinity, minY: Infinity, maxY: -Infinity });
}

function verticalRangeAtX(points, x) {
  const ys = [];
  for (let i = 0; i < points.length; i += 1) {
    const [x1, y1] = points[i];
    const [x2, y2] = points[(i + 1) % points.length];
    const minX = Math.min(x1, x2);
    const maxX = Math.max(x1, x2);
    if (x < minX || x > maxX) continue;
    if (x1 === x2) {
      ys.push(y1, y2);
      continue;
    }
    const t = (x - x1) / (x2 - x1);
    if (t >= 0 && t <= 1) ys.push(y1 + (y2 - y1) * t);
  }
  if (ys.length < 2) {
    const bounds = polygonBounds(points);
    return { min: bounds.minY, max: bounds.maxY };
  }
  ys.sort((a, b) => a - b);
  return { min: ys[0], max: ys[ys.length - 1] };
}

function clampPointToGround(x, y) {
  const points = currentWalkPolygon();
  const bounds = polygonBounds(points);
  const groundX = clamp(x, bounds.minX, bounds.maxX);
  const range = verticalRangeAtX(points, groundX);
  return {
    x: groundX,
    y: clamp(y, range.min, range.max),
  };
}

function randomGroundPoint() {
  const points = currentWalkPolygon();
  const bounds = polygonBounds(points);
  for (let attempt = 0; attempt < 12; attempt += 1) {
    const x = bounds.minX + Math.random() * (bounds.maxX - bounds.minX);
    const range = verticalRangeAtX(points, x);
    if (range.max - range.min < 8) continue;
    return {
      x,
      y: range.min + 6 + Math.random() * Math.max(1, range.max - range.min - 12),
    };
  }
  return catPoint("center");
}

function setActorTarget(point, state = actor.state, emote = "") {
  const grounded = clampPointToGround(point.x, point.y);
  actor.targetX = grounded.x;
  actor.targetY = grounded.y;
  actor.state = state;
  if (emote) showEmote(emote, 1.8);
}

function showEmote(emote, seconds = 1.8) {
  actor.emote = emote;
  actor.emoteClock = Math.max(actor.emoteClock, seconds);
}

function currentCatMood(room) {
  if (!model || !room) return "patrol";
  if (model.state.paused || model.state.mining_paused) return "reboot";
  if (room.heat_pct >= 0.95) return "critical";
  if (room.heat_pct >= 0.8) return "hot";
  return "patrol";
}

function eventReaction(event) {
  if (!event) return null;
  const byID = {
    petty_thief: { state: "alarm", point: "door", emote: "!", seconds: 4.2 },
    admin_discovery: { state: "alarm", point: "door", emote: "!", seconds: 4.6 },
    tunnel_heist: { state: "alarm", point: "door", emote: "!!", seconds: 5.0 },
    police_visit: { state: "alarm", point: "door", emote: "!!", seconds: 5.0 },
    government_raid: { state: "alarm", point: "door", emote: "!!", seconds: 5.4 },
    tax_audit: { state: "confused", point: "door", emote: "?", seconds: 4.0 },
    power_outage: { state: "reboot", point: "center", emote: "...", seconds: 3.8 },
    gas_leak: { state: "alarm", point: "door", emote: "!", seconds: 4.2 },
    voltage_dip: { state: "inspect", point: "rig", emote: "⚡", seconds: 3.8 },
    power_surge: { state: "alarm", point: "rig", emote: "⚡!", seconds: 4.4 },
    rain_leak: { state: "alarm", point: "rig", emote: "💧", seconds: 4.0 },
    tech_share: { state: "happy", point: "rest", emote: "💡", seconds: 3.6 },
    extra_delivery: { state: "happy", point: "door", emote: "📦", seconds: 3.8 },
    btc_pump: { state: "happy", point: "rig", emote: "₿", seconds: 4.0 },
    lucky_fish: { state: "happy", point: "rest", emote: "♪", seconds: 4.0 },
    street_dog: { state: "confused", point: "door", emote: "?", seconds: 3.6 },
    group_chat_sos: { state: "confused", point: "rest", emote: "💬", seconds: 3.6 },
  };
  if (byID[event.id]) return byID[event.id];
  if (event.category === "crisis") return { state: "alarm", point: "door", emote: "!!", seconds: 5.0 };
  if (event.category === "threat") return { state: "alarm", point: "rig", emote: "!", seconds: 4.0 };
  if (event.category === "opportunity") return { state: "happy", point: "rig", emote: "♪", seconds: 3.6 };
  if (event.category === "social") return { state: "confused", point: "door", emote: "?", seconds: 3.4 };
  return { state: "inspect", point: "center", emote: "...", seconds: 2.8 };
}

function distanceToTarget() {
  return Math.hypot(actor.targetX - actor.x, actor.targetY - actor.y);
}

function coolingPoint(fan, mood) {
  const offsets = mood === "critical"
    ? [[-20, 0], [16, 4], [-4, 10]]
    : [[-16, 0], [16, 3], [0, 9]];
  const [dx, dy] = offsets[actor.coolingIndex % offsets.length];
  actor.coolingIndex += 1;
  return { x: fan.x + dx, y: fan.y + dy };
}

function handleCatAI(dt) {
  if (!model) return;
  const room = currentRoom();
  if (!room) return;

  actor.stateClock += dt;
  actor.decisionClock -= dt;
  actor.reactionClock = Math.max(0, actor.reactionClock - dt);
  actor.emoteClock = Math.max(0, actor.emoteClock - dt);

  const event = model.last_event;
  const eventKey = event ? `${event.seq || 0}:${event.id}` : "";
  if (event && eventKey !== lastEventKey) {
    lastEventKey = eventKey;
    const reaction = eventReaction(event);
    if (reaction) {
      actor.reactionClock = reaction.seconds;
      actor.decisionClock = reaction.seconds;
      setActorTarget(catPoint(reaction.point), reaction.state, reaction.emote);
      return;
    }
  }

  const mood = currentCatMood(room);
  if (mood === "reboot") {
    actor.state = "reboot";
    actor.targetX = actor.x;
    actor.targetY = actor.y;
    actor.moving = false;
    if (actor.emoteClock <= 0) showEmote("...", 1.5);
    return;
  }
  if ((mood === "hot" || mood === "critical") && actor.reactionClock <= 0) {
    const fan = catPoint("fan");
    const nearFan = Math.hypot(actor.x - fan.x, actor.y - fan.y) < 26;
    const arrived = distanceToTarget() <= ARRIVAL_DISTANCE;
    if (actor.state !== mood || (!nearFan && actor.decisionClock <= 0)) {
      setActorTarget(coolingPoint(fan, mood), mood, mood === "critical" ? "!!" : "热");
      actor.decisionClock = mood === "critical" ? 2.8 : 3.8;
    } else if (arrived && actor.decisionClock <= 0) {
      const next = coolingPoint(fan, mood);
      const nextDistance = Math.hypot(next.x - actor.x, next.y - actor.y);
      if (nextDistance > 12) {
        setActorTarget(next, mood, mood === "critical" ? "!!" : "热");
      } else {
        actor.direction = "down";
        if (actor.emoteClock <= 0) showEmote(mood === "critical" ? "!!" : "热", mood === "critical" ? 1.3 : 1.7);
      }
      actor.decisionClock = mood === "critical" ? 2.6 : 3.6;
    } else if (arrived) {
      actor.targetX = actor.x;
      actor.targetY = actor.y;
      actor.direction = "down";
      if (actor.emoteClock <= 0) showEmote(mood === "critical" ? "!!" : "热", mood === "critical" ? 1.3 : 1.7);
    }
    return;
  }
  if (actor.reactionClock > 0) return;

  if (actor.decisionClock > 0) return;

  const gpusInRoom = (model.gpus || []).filter((gpu) => gpu.room === model.state.current_room);
  const roll = Math.random();
  if (gpusInRoom.length && roll < 0.38) {
    const rig = catPoint("rig");
    setActorTarget({ x: rig.x + (Math.random() - 0.5) * 42, y: rig.y }, "inspect", "...");
    actor.decisionClock = 2.2 + Math.random() * 2.0;
  } else if (roll < 0.56) {
    setActorTarget(catPoint("rest"), "rest");
    actor.decisionClock = 1.4 + Math.random() * 2.4;
  } else {
    setActorTarget(randomGroundPoint(), "patrol");
    actor.decisionClock = 1.8 + Math.random() * 3.0;
  }
}

function movementVector() {
  const toX = actor.targetX - actor.x;
  const toY = actor.targetY - actor.y;
  const distance = Math.hypot(toX, toY);
  if (distance <= ARRIVAL_DISTANCE) {
    actor.targetX = actor.x;
    actor.targetY = actor.y;
    pointer.active = false;
    return { dx: 0, dy: 0 };
  }
  return { dx: toX / distance, dy: toY / distance };
}

function setDirection(dx, dy) {
  if (Math.abs(dx) > Math.abs(dy)) actor.direction = dx < 0 ? "left" : "right";
  else if (Math.abs(dy) > 0.01) actor.direction = dy < 0 ? "up" : "down";
}

function updateActor(dt) {
  handleCatAI(dt);
  const { dx, dy } = movementVector();
  actor.moving = Math.abs(dx) + Math.abs(dy) > 0;
  if (actor.moving) {
    setDirection(dx, dy);
    actor.x += dx * actorSpeed() * dt;
    actor.y += dy * actorSpeed() * dt;
  }
  const grounded = clampPointToGround(actor.x, actor.y);
  actor.x = grounded.x;
  actor.y = grounded.y;
  const frameDuration = actor.moving ? 0.12 : 0.24;
  actor.frameClock += dt;
  if (actor.frameClock >= frameDuration) {
    actor.frameClock = 0;
    actor.frame = (actor.frame + 1) % FRAME_COUNT;
  }
  actor.bobClock += dt;
}

function actorSpeed() {
  if (actor.state === "alarm") return 118;
  if (actor.state === "critical") return 74;
  if (actor.state === "hot") return 58;
  if (actor.state === "happy") return 82;
  if (actor.state === "called") return 76;
  if (actor.state === "inspect") return 52;
  if (actor.state === "rest") return distanceToTarget() > 5 ? 42 : 0;
  if (actor.state === "reboot") return 0;
  return 46;
}

function drawRigs(room) {
  // The room art already contains mining rigs; exact inventory lives in the UI.
}

function drawPlayer() {
  const mode = actor.moving ? "walk" : "idle";
  const sprite = images.sprites.get(`${mode}:${actor.direction}:${actor.frame + 1}`);
  ctx.save();
  ctx.globalAlpha = 0.32;
  ctx.fillStyle = "#030607";
  ctx.beginPath();
  ctx.ellipse(actor.x, actor.y + 1, 15 * actor.scale, 4 * actor.scale, 0, 0, Math.PI * 2);
  ctx.fill();
  ctx.restore();
  if (!sprite) return;
  const { image, bounds } = sprite;
  const visibleHeight = actor.targetBodyHeight * actor.scale;
  const drawScale = visibleHeight / bounds.height;
  const w = bounds.width * drawScale;
  const h = bounds.height * drawScale;
  ctx.drawImage(
    image,
    bounds.x,
    bounds.y,
    bounds.width,
    bounds.height,
    Math.round(actor.x - w / 2),
    Math.round(actor.y - h),
    Math.round(w),
    Math.round(h),
  );
}

function drawEmote() {
  if (!actor.emote || actor.emoteClock <= 0) return;
  const alpha = Math.min(1, actor.emoteClock / 0.35);
  const text = actor.emote;
  const x = Math.round(actor.x + 18);
  const y = Math.round(actor.y - actor.targetBodyHeight * actor.scale - 12);
  ctx.save();
  ctx.globalAlpha = alpha;
  ctx.font = "bold 11px monospace";
  const width = Math.max(18, ctx.measureText(text).width + 12);
  ctx.fillStyle = "rgba(7, 12, 14, 0.86)";
  ctx.strokeStyle = actor.state === "alarm" || actor.state === "critical" ? "#ff814d" : "#70e39a";
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.roundRect(x, y - 16, width, 18, 5);
  ctx.fill();
  ctx.stroke();
  ctx.fillStyle = actor.state === "alarm" || actor.state === "critical" ? "#ffc857" : "#dfffe9";
  ctx.fillText(text, x + 6, y - 3);
  ctx.restore();
}

function draw() {
  ctx.clearRect(0, 0, STAGE.width, STAGE.height);
  const room = currentRoom();
  const bg = room ? images.rooms.get(room.id) : null;
  if (bg) ctx.drawImage(bg, 0, 0, STAGE.width, STAGE.height);
  else {
    ctx.fillStyle = "#071013";
    ctx.fillRect(0, 0, STAGE.width, STAGE.height);
  }
  drawRigs(room);
  if (pointer.active) {
    ctx.save();
    ctx.strokeStyle = "#70e39a";
    ctx.beginPath();
    ctx.arc(pointer.x, pointer.y, 5, 0, Math.PI * 2);
    ctx.moveTo(pointer.x - 8, pointer.y);
    ctx.lineTo(pointer.x + 8, pointer.y);
    ctx.moveTo(pointer.x, pointer.y - 8);
    ctx.lineTo(pointer.x, pointer.y + 8);
    ctx.stroke();
    ctx.restore();
  }
  drawPlayer();
  drawEmote();
  if (model?.state.paused || model?.state.mining_paused) {
    ctx.save();
    ctx.fillStyle = "rgba(5, 8, 10, 0.42)";
    ctx.fillRect(0, 0, STAGE.width, STAGE.height);
    ctx.fillStyle = "#ffc857";
    ctx.font = "12px monospace";
    ctx.fillText(model.state.paused ? "PAUSED" : "REBOOTING", 18, 28);
    ctx.restore();
  }
}

function loop(now) {
  const dt = Math.min(0.05, (now - previousTime) / 1000);
  previousTime = now;
  updateActor(dt);
  draw();
  requestAnimationFrame(loop);
}

function canvasPoint(event) {
  const rect = canvas.getBoundingClientRect();
  return {
    x: ((event.clientX - rect.left) / rect.width) * STAGE.width,
    y: ((event.clientY - rect.top) / rect.height) * STAGE.height,
  };
}

window.addEventListener("keydown", (event) => {
  const key = event.key.length === 1 ? event.key.toLowerCase() : event.key;
  if (["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown", "w", "a", "s", "d"].includes(key)) {
    keys.add(key);
  }
});

window.addEventListener("keyup", (event) => {
  const key = event.key.length === 1 ? event.key.toLowerCase() : event.key;
  keys.delete(key);
});

canvas.addEventListener("pointerdown", (event) => {
  const p = canvasPoint(event);
  const grounded = clampPointToGround(p.x, p.y);
  pointer.x = grounded.x;
  pointer.y = grounded.y;
  pointer.active = true;
  setActorTarget(grounded, "called", "?");
  actor.decisionClock = 2.4;
  canvas.setPointerCapture(event.pointerId);
});

canvas.addEventListener("pointermove", (event) => {
  if (!pointer.active || event.buttons === 0) return;
  const p = canvasPoint(event);
  const grounded = clampPointToGround(p.x, p.y);
  pointer.x = grounded.x;
  pointer.y = grounded.y;
  setActorTarget(grounded, "called");
  actor.decisionClock = 2.4;
});

pauseButton.addEventListener("click", () => action({ action: "toggle_pause" }));
ventButton.addEventListener("click", () => action({ action: "vent" }));
resetButton.addEventListener("click", () => action({ action: "reset" }));
window.addEventListener("resize", syncSideHeight);
new ResizeObserver(syncSideHeight).observe(stageShell);

initTabs();
Promise.all([preloadSprites(), loadSceneMap()])
  .then(fetchSnapshot)
  .then(() => {
    previousTime = performance.now();
    requestAnimationFrame(loop);
    setInterval(fetchSnapshot, 1000);
  })
  .catch((error) => {
    lastMessage = error.message;
    toast.textContent = error.message;
  });
