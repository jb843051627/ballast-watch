// ballast-watch 看板逻辑：轮询 /api/v1/dashboard 渲染房间卡片。
(function () {
  "use strict";

  const tanksEl = document.getElementById("tanks");
  const tankCountEl = document.getElementById("tank-count");
  const compliance_alertCountEl = document.getElementById("compliance_alert-count");
  const cacheTimeEl = document.getElementById("cache-time");
  const updatedAtEl = document.getElementById("updated-at");

  const STATUS_TEXT = {
    at_rest: "静态",
    normal: "正常",
    compliance_alert: "告警",
    alarm: "严重",
    restricted: "受限",
    release: "放行",
  };

  const PARAM_TEXT = {
    temp: "温度(℃)",
    humidity: "湿度(%)",
    pressure: "压差(Pa)",
    particle_05: "粒子≥0.5μm",
    particle_50: "粒子≥5.0μm",
  };

  async function refresh() {
    try {
      const resp = await fetch("/api/v1/dashboard");
      if (!resp.ok) throw new Error("HTTP " + resp.status);
      const data = await resp.json();
      render(data);
    } catch (e) {
      updatedAtEl.textContent = "加载失败: " + e.message;
    }
  }

  function render(data) {
    updatedAtEl.textContent = new Date(data.updated_at).toLocaleString("zh-CN");
    tankCountEl.textContent = data.tanks ? data.tanks.length : 0;
    compliance_alertCountEl.textContent = data.open_compliance_compliance_alerts || 0;
    cacheTimeEl.textContent = data.updated_at ? new Date(data.updated_at).toLocaleTimeString("zh-CN") : "-";

    const tanks = data.tanks || [];
    tanksEl.innerHTML = "";
    for (const tank of tanks) {
      tanksEl.appendChild(renderBallastTank(tank));
    }
  }

  function renderBallastTank(tank) {
    const div = document.createElement("div");
    div.className = "tank status-" + (tank.status || "normal");

    const head = document.createElement("div");
    head.className = "tank-head";

    const title = document.createElement("h3");
    title.textContent = tank.tank_code + " " + (STATUS_TEXT[tank.status] || tank.status);
    head.appendChild(title);

    const badge = document.createElement("span");
    badge.className = "badge" + (tank.status === "alarm" ? " alarm" : tank.status === "compliance_alert" ? " compliance_alert" : "");
    badge.textContent = "未决 " + tank.open_compliance_compliance_alerts;
    head.appendChild(badge);

    div.appendChild(head);

    const water_readings = document.createElement("div");
    water_readings.className = "water_readings";
    const rt = tank.realtime || [];
    if (rt.length === 0) {
      const empty = document.createElement("div");
      empty.className = "empty";
      empty.textContent = "暂无读数";
      water_readings.appendChild(empty);
    } else {
      for (const r of rt) {
        const row = document.createElement("div");
        row.className = "water_reading";
        const name = document.createElement("span");
        name.textContent = PARAM_TEXT[r.param_type] || r.param_type;
        const val = document.createElement("span");
        val.className = "val " + (r.within_range ? "ok" : "bad");
        val.textContent = r.value.toFixed(2);
        row.appendChild(name);
        row.appendChild(val);
        water_readings.appendChild(row);
      }
    }
    div.appendChild(water_readings);
    return div;
  }

  document.getElementById("refresh-btn").addEventListener("click", refresh);
  refresh();
  setInterval(refresh, 5000);
})();