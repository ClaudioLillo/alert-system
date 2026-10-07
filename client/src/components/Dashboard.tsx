interface DeviceData {
    id: string;
    createdAt: string;
    isActive: boolean;
    lastSeenAt?: string;
    name: string;
    updatedAt: string;


}

interface DashboardProps {
    devices: DeviceData[];
}

export default function Dashboard({devices}: DashboardProps) {
    return (
        <div className="dashboard">
            {devices.map((device) => (
                <div key={device.id} className="device-card">
                    <span>{device.name}</span>
                    <span>ID: {device.id}</span>
                </div>
            ))}
        </div>
    )

}