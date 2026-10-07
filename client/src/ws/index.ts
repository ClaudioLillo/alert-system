import { v4 as uuidv4 } from 'uuid';

export default class DeviceConnection {
    ws: WebSocket;
    deviceId: string;

    constructor() {
        this.ws = new WebSocket('ws://localhost:4000');
        this.deviceId = uuidv4();
    }
}