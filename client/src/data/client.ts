import axios from 'axios';

const config ={
  baseURL: 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
};

const client = axios.create(config);

export default client;