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

// 添加请求拦截器以便调试
client.interceptors.request.use(request => {
  console.log('开始请求:', request.method?.toUpperCase(), request.url);
  console.log('请求数据:', request.data);
  return request;
});

client.interceptors.response.use(response => {
  console.log('响应:', response.status, response.config.url);
  return response;
}, error => {
  console.log('响应错误:', error.message);
  if (error.response) {
    console.log('错误数据:', error.response.data);
    console.log('错误状态:', error.response.status);
  }
  return Promise.reject(error);
});

export default client;
