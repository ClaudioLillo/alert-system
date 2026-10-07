import { useEffect, useState } from 'react'
import './App.css'
import DeviceConnection from './ws'
import Device from './components/Device'
import { getDevices } from './data/devices'
import Dashboard from './components/Dashboard'

function App() {
  const [devices, setDevices] = useState<DeviceConnection[]>([])
  const [devicesData, setDevicesData] = useState<any[]>([])

  const addDevice = () => {
    const newDevice = new DeviceConnection();
    setDevices([...devices, newDevice]);
  }

  useEffect(() => {
    getDevices().then((data) => {
      setDevicesData(data)
    })
  })

  return (
    <div className="app">
      <div className="left-panel">
      <div className="button" onClick={addDevice}>
        Agregar Dispositivo
      </div>
      <div className="devices">
      {devices.map((device, index) => (
        <Device
          key={index}
          deviceId={device.deviceId}
          ws={device.ws}
        />
      ))}
      </div>
      </div>
      <div className="right-panel">
        <h2>Dispositivos Registrados</h2>
        <Dashboard devices={devicesData} />
      </div>
    </div>
  )
}

export default App
