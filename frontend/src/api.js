import axios from 'axios'

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
export const WS_BASE_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080'

const api = axios.create({
  baseURL: `${API_BASE_URL}/api`,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

export function getErrorMessage(err) {
  return err?.response?.data?.error || err?.message || 'Something went wrong'
}

export default api
