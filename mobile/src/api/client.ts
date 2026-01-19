import axios from 'axios';
import { Platform } from 'react-native';

const BASE_URL = Platform.OS === 'android' ? 'http://10.0.2.2:8080' : 'http://localhost:8080';

const client = axios.create({
  baseURL: BASE_URL,
  timeout: 60000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor for debugging
client.interceptors.request.use(request => {
  console.log('Starting Request:', request.method?.toUpperCase(), request.url);
  console.log('Request Data:', request.data);
  return request;
});

client.interceptors.response.use(response => {
  console.log('Response:', response.status, response.config.url);
  return response;
}, error => {
  console.log('Response Error:', error.message);
  if (error.response) {
      console.log('Error Data:', error.response.data);
      console.log('Error Status:', error.response.status);
  }
  return Promise.reject(error);
});

export default client;
