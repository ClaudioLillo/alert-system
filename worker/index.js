import { Client } from 'pg';
import 'dotenv/config';
import { WebSocketServer } from 'ws';

const config = {
        user: process.env.DB_USER,
        host: process.env.DB_HOST,
        database: process.env.DB_NAME,
        password: process.env.DB_PASSWORD,
        port: Number(process.env.DB_PORT),
        ssl: {
            rejectUnauthorized: false
        }
    }

const wss = new WebSocketServer({ port: 4000 });

wss.on('connection', (ws) => {
    console.log('Client connected:', ws.id);

    ws.on('message', (message) => {
        console.log('Received message:', message.toString());
    });
});

const start = async() => {
    const db = new Client(config);
    await db.connect();
    await db.query('LISTEN new_alert');

    db.on('notification', (msg) => {
        const {id, device_id, status, created_at} = JSON.parse(msg.payload);
        console.log('New alert received:', {id, device_id, status, created_at});
        wss.clients.forEach((client)=> {
            if (client.readyState === 1) {
                console.log('Sending alert to client:', {id, device_id, status, created_at});
                // client.send(JSON.stringify({id, device_id, status, created_at}));
            }
        })
    });

    db.on('error', (err) => {
        console.error('Database error:', err);
    });
}

start();
