import axios from "axios";

const api = axios.create({
  baseURL: "http://localhost:8080", // or your backend base URL
});

// Set the Authorization header for every request if token exists
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;
