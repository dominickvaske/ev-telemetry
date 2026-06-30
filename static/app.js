const POLL_MS = 1000;

function batteryBar(pct) {
  const color = pct > 50 ? 'green' : pct > 20 ? 'yellow' : 'red';
  const filled = Math.round(pct / 10);
  let segs = '';
  for (let i = 0; i < 10; i++) {
    segs += `<div class="seg ${i < filled ? 'on ' + color : ''}"></div>`;
  }
  return `
    <div class="battery-cell">
      <div class="battery-bar">${segs}</div>
      <span class="battery-pct ${color}">${pct.toFixed(1)}%</span>
    </div>`;
}

function statusBadge(vehicle) {
  if (vehicle.is_charging) return '<span class="badge charging">⚡ CHARGING</span>';
  if (vehicle.battery_pct < 10) return '<span class="badge critical">CRITICAL</span>';
  if (vehicle.battery_pct < 20) return '<span class="badge warn">LOW BAT</span>';
  if (vehicle.temp_c > 50) return '<span class="badge warn">HIGH TEMP</span>';
  if (vehicle.speed_kph > 100) return '<span class="badge warn">OVERSPEED</span>';
  return '<span class="badge ok">OK</span>';
}

function formatTime(ts) {
  return new Date(ts).toLocaleTimeString();
}

function alertLabel(type) {
  if (type === 'BATTERY') return 'LOW BATTERY';
  if (type === 'TEMPERATURE') return 'HIGH TEMP';
  return type;
}

async function fetchJSON(url) {
  const res = await fetch(url);
  return res.json();
}

async function update() {
  try {
    const [fleet, alerts, status] = await Promise.all([
      fetchJSON('/fleet'),
      fetchJSON('/alerts'),
      fetchJSON('/status'),
    ]);

    // Header
    document.getElementById('uptime').textContent = status.uptime;
    document.getElementById('events').textContent = status.events_total.toLocaleString();
    document.getElementById('goroutines').textContent = status.goroutines;
    document.getElementById('port').textContent = ':' + status.port;

    // Summary bar
    const charging = fleet.filter(v => v.is_charging).length;
    const avgBat = fleet.length
      ? (fleet.reduce((s, v) => s + v.battery_pct, 0) / fleet.length).toFixed(1)
      : 0;

    document.getElementById('fleet-count').textContent = fleet.length;
    document.getElementById('avg-battery').textContent = avgBat + '%';
    document.getElementById('charging-count').textContent = charging;
    document.getElementById('alert-count').textContent = alerts.length;

    // Fleet table — sort by ID for stable ordering
    fleet.sort((a, b) => a.id.localeCompare(b.id));
    const tbody = document.getElementById('fleet-body');
    tbody.innerHTML = fleet.map(v => `
      <tr>
        <td class="vehicle-id">${v.id}</td>
        <td>${batteryBar(v.battery_pct)}</td>
        <td>${v.speed_kph.toFixed(0)} km/h</td>
        <td>${v.temp_c.toFixed(1)}°C</td>
        <td>${statusBadge(v)}</td>
      </tr>`).join('');

    // Alert log — newest first, cap at 50
    const recent = [...alerts].reverse().slice(0, 50);
    document.getElementById('alert-list').innerHTML = recent.map(a => `
      <div class="alert-card">
        <div class="alert-card-header">
          <span class="alert-type ${a.type}">${alertLabel(a.type)}</span>
          <span class="alert-time">${formatTime(a.timestamp)}</span>
        </div>
        <div class="alert-msg">
          <span class="alert-vehicle">${a.vehicle_id}</span> — ${a.message}
        </div>
      </div>`).join('');

  } catch (err) {
    console.error('fetch error:', err);
  }
}

update();
setInterval(update, POLL_MS);
