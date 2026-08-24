import axios from 'axios'
import { env } from './env'

const api = axios.create({
  baseURL: env.apiBaseUrl,
  withCredentials: true,
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    return Promise.reject(error)
  }
)

export default api
