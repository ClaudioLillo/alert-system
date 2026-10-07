import { useEffect } from 'react';
import './Device.css';

interface DeviceProps {
    deviceId: string;
    ws: WebSocket;
}

export default function Device({ deviceId, ws }: DeviceProps) {
    useEffect(() => {
        ws.onopen = () => {
            console.log('Conexión WebSocket establecida para el dispositivo:', deviceId);
        }
    })
    return (
        <div className="device">
            <p>Device ID: {deviceId.split("-")[0]}</p>
        </div>
    );
}