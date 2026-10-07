import client from './client';

export const getDevices = async () => {
    const data =await client.get('/devices');
    return data.data;
}