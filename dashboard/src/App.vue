<template>
  <div id="app-container">
    <div id="header">
      <h1>OptiRoute — 1-Hour Demand Forecast</h1>
      <span id="last-updated">Last updated: {{ lastUpdated }}</span>
    </div>
    <div id="main-content">
      <div id="map-container">
        <div id="map"></div>
      </div>
      <div id="sidebar">
        <h2>Zone Info</h2>
        <div v-if="selectedZone">
          <h3>{{ selectedZone.name }}</h3>
          <p>30m forecast: <strong>{{ selectedZone.predicted }} orders</strong></p>
          <p>Current demand: <strong>{{ selectedZone.currentOrders }} orders</strong></p>
          <div id="shap-section">
            <h4>Why this demand?</h4>
            <div v-for="factor in selectedZone.factors" :key="factor.feature">
              <span class="factor-name">{{ factor.feature }}</span>
              <span class="factor-value" :class="factor.contribution > 0 ? 'positive' : 'negative'">
                {{ factor.contribution > 0 ? '+' : '' }}{{ factor.contribution }}
              </span>
            </div>
          </div>
        </div>
        <div v-else>
          <p>Click a zone circle to zoom in and inspect demand.</p>
        </div>
        <button class="demo-button" @click="seedDemoFleet" :disabled="isSeedingFleet">
          {{ isSeedingFleet ? 'Spawning demo fleet...' : 'Spawn demo fleet' }}
        </button>
        <div id="active-drivers">
          <h3>Active Drivers: {{ activeDriverCount }}</h3>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

const lastUpdated = ref('Loading...')
const selectedZone = ref(null)
const activeDriverCount = ref(0)
const isSeedingFleet = ref(false)

const ZONES = [
  { name: 'koramangala',     lat: 12.9352, lng: 77.6245 },
  { name: 'indiranagar',     lat: 12.9784, lng: 77.6408 },
  { name: 'whitefield',      lat: 12.9698, lng: 77.7499 },
  { name: 'marathahalli',    lat: 12.9591, lng: 77.6974 },
  { name: 'hsr_layout',      lat: 12.9116, lng: 77.6389 },
  { name: 'jp_nagar',        lat: 12.9102, lng: 77.5856 },
  { name: 'electronic_city', lat: 12.8399, lng: 77.6770 },
  { name: 'hebbal',          lat: 13.0353, lng: 77.5972 },
]

const DEMO_DRIVERS = [
  { id: 'driver_001', lat: 12.9352, lng: 77.6245 },
  { id: 'driver_002', lat: 12.9784, lng: 77.6408 },
  { id: 'driver_003', lat: 12.9102, lng: 77.5856 },
  { id: 'driver_004', lat: 12.9698, lng: 77.7499 },
  { id: 'driver_005', lat: 12.8399, lng: 77.6770 },
  { id: 'driver_006', lat: 12.9591, lng: 77.6974 },
  { id: 'driver_007', lat: 12.9116, lng: 77.6389 },
  { id: 'driver_008', lat: 13.0353, lng: 77.5972 },
  { id: 'driver_009', lat: 12.9485, lng: 77.6157 },
  { id: 'driver_010', lat: 12.9869, lng: 77.6457 },
  { id: 'driver_011', lat: 12.9253, lng: 77.7065 },
  { id: 'driver_012', lat: 12.8627, lng: 77.6692 },
  { id: 'driver_013', lat: 12.9964, lng: 77.5600 },
  { id: 'driver_014', lat: 12.9178, lng: 77.6382 },
  { id: 'driver_015', lat: 12.9337, lng: 77.5898 },
  { id: 'driver_016', lat: 12.9767, lng: 77.6024 },
  { id: 'driver_017', lat: 12.9045, lng: 77.6460 },
  { id: 'driver_018', lat: 12.9458, lng: 77.6712 },
  { id: 'driver_019', lat: 13.0083, lng: 77.6255 },
  { id: 'driver_020', lat: 12.8806, lng: 77.5931 },
]

let map = null
const zoneLayers = {}
const driverMarkers = []

const apiBase = 'http://localhost:8080'
const baseOrders = 10
const defaultView = {
  center: [12.955, 77.63],
  zoom: 12,
}

const getDemandColor = (predicted) => {
  if (predicted >= 25) return '#E24B4A'
  if (predicted >= 18) return '#BA7517'
  if (predicted >= 12) return '#F5C842'
  return '#1D9E75'
}

const getZoneStyle = (predicted) => {
  const intensity = Math.max(0.18, Math.min(0.45, predicted / 80))
  const heatRadius = 2200 + (predicted * 140)
  const ringRadius = 10 + (predicted / 4)

  return {
    color: getDemandColor(predicted),
    heatRadius,
    heatOpacity: intensity,
    ringRadius,
  }
}

const updateSelectedZone = (zone, data) => {
  selectedZone.value = {
    name: zone.name,
    predicted: data.predicted_orders,
    currentOrders: baseOrders,
    factors: data.top_factors,
  }
}

const focusZone = (zone) => {
  if (!map) return

  map.flyTo([zone.lat, zone.lng], 14, {
    duration: 0.8,
  })
}

const putDriverLocation = async (driver) => {
  const jitter = () => ((Math.random() - 0.5) * 0.0025)
  const latitude = Number((driver.lat + jitter()).toFixed(4))
  const longitude = Number((driver.lng + jitter()).toFixed(4))

  await fetch(`${apiBase}/api/v1/drivers/location`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      driver_id: driver.id,
      latitude,
      longitude,
    }),
  })
}

const seedDemoFleet = async () => {
  if (isSeedingFleet.value) return

  isSeedingFleet.value = true

  try {
    await Promise.all(DEMO_DRIVERS.map((driver) => putDriverLocation(driver)))
    await fetchActiveDrivers()
  } catch (err) {
    console.log('Could not seed demo fleet:', err)
  } finally {
    isSeedingFleet.value = false
  }
}

const initMap = () => {
  map = L.map('map', {
    preferCanvas: true,
  }).setView(defaultView.center, defaultView.zoom)
  L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
    attribution: '© OpenStreetMap © CARTO',
    subdomains: 'abcd',
    maxZoom: 19
  }).addTo(map)
}

const fetchPredictions = async () => {
  const now = new Date()
  // Forecast window: predict demand for the next hour
  const forecastHour = (now.getHours() + 1) % 24
  const day = now.getDay() === 0 ? 6 : now.getDay() - 1
  const preservedCenter = map ? map.getCenter() : null
  const preservedZoom = map ? map.getZoom() : null

  for (const zone of ZONES) {
    try {
      const response = await fetch('http://localhost:8000/api/v1/predict_demand', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          zone_id: zone.name,
          hour_of_day: forecastHour,
          day_of_week: day,
          is_raining: false,
          current_orders: baseOrders
        })
      })
      const data = await response.json()
      const { color, heatRadius, heatOpacity, ringRadius } = getZoneStyle(data.predicted_orders)

      if (zoneLayers[zone.name]) {
        map.removeLayer(zoneLayers[zone.name].heat)
        map.removeLayer(zoneLayers[zone.name].ring)
      }

      const heat = L.circle([zone.lat, zone.lng], {
        color: color,
        fillColor: color,
        fillOpacity: heatOpacity,
        opacity: 0.2,
        weight: 1,
        radius: heatRadius
      }).addTo(map)

      const ring = L.circleMarker([zone.lat, zone.lng], {
        radius: ringRadius,
        color: '#F2EFFF',
        weight: 2,
        opacity: 0.9,
        fillColor: color,
        fillOpacity: 0.95,
      }).addTo(map)

      const popupHtml = `<b>${zone.name}</b><br>Forecast: ${data.predicted_orders} orders / 30m<br>Current: ${baseOrders} orders`
      heat.bindPopup(popupHtml)
      ring.bindPopup(popupHtml)

      const onZoneClick = () => {
        updateSelectedZone(zone, data)
        focusZone(zone)
      }

      heat.on('click', onZoneClick)
      ring.on('click', onZoneClick)

      zoneLayers[zone.name] = {
        heat,
        ring,
      }
    } catch (err) {
      console.log(`Could not fetch prediction for ${zone.name}:`, err)
    }
  }

  if (preservedCenter && preservedZoom !== null) {
    map.setView(preservedCenter, preservedZoom, { animate: false })
  }

  lastUpdated.value = new Date().toLocaleTimeString()
}

const fetchActiveDrivers = async () => {
  try {
    const response = await fetch(`${apiBase}/api/v1/drivers/active`)
    const data = await response.json()

    driverMarkers.forEach(m => map.removeLayer(m))
    driverMarkers.length = 0

    if (data.drivers) {
      activeDriverCount.value = data.count
      data.drivers.forEach(driver => {
        const marker = L.circleMarker([driver.latitude, driver.longitude], {
          radius: 5,
          color: '#8A7CF8',
          weight: 1,
          fillColor: '#8A7CF8',
          fillOpacity: 0.85,
          opacity: 0.85,
        }).addTo(map)
        marker.bindPopup(`Driver: ${driver.driver_id}`)
        driverMarkers.push(marker)
      })
    }
  } catch (err) {
    console.log('Could not fetch active drivers:', err)
  }
}

const refreshAll = async () => {
  await fetchPredictions()
  await fetchActiveDrivers()
}

onMounted(async () => {
  initMap()
  await refreshAll()
  if (activeDriverCount.value < 10) {
    await seedDemoFleet()
  }
  setInterval(refreshAll, 30000)
})
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: Arial, sans-serif; background: #0C1020; color: #E0E0FF; }

#app-container { display: flex; flex-direction: column; height: 100vh; }

#header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: #16213E;
  border-bottom: 1px solid #7F77DD;
}

#header h1 { font-size: 18px; color: #E7E4FF; }
#last-updated { font-size: 12px; color: #7A7A9A; }

#main-content { display: flex; flex: 1; overflow: hidden; }
#map-container { flex: 1; }
#map { width: 100%; height: 100%; }

#sidebar {
  width: 280px;
  background: #16213E;
  padding: 16px;
  overflow-y: auto;
  border-left: 1px solid #7F77DD;
}

#sidebar h2 { font-size: 14px; color: #7F77DD; margin-bottom: 12px; }
#sidebar h3 { font-size: 13px; color: #C0BCFF; margin-bottom: 8px; }
#sidebar h4 { font-size: 11px; color: #9A9AB0; margin-bottom: 6px; }
#sidebar p  { font-size: 12px; color: #9A9AB0; line-height: 1.6; }

.factor-name  { font-size: 11px; color: #9A9AB0; display: inline-block; width: 160px; }
.factor-value { font-size: 11px; font-weight: bold; }
.positive { color: #E24B4A; }
.negative { color: #1D9E75; }

#shap-section {
  margin: 10px 0;
  padding: 10px;
  background: #1A1A2E;
  border-radius: 6px;
}

#active-drivers {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #2A2A4E;
}

.demo-button {
  width: 100%;
  margin-top: 14px;
  padding: 10px 12px;
  border: 0;
  border-radius: 8px;
  background: linear-gradient(135deg, #7F77DD, #50C9A9);
  color: #08111F;
  font-weight: 700;
  cursor: pointer;
}

.demo-button:disabled {
  opacity: 0.7;
  cursor: progress;
}
</style>