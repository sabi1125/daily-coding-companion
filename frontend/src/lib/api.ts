import axios from 'axios'
import { env } from './env'

const api = axios.create({
  baseURL: env.apiBaseUrl,
  withCredentials: true,
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status == 401) {
      window.location.assign('/')
    }
    return Promise.reject(error)
  }
)

export default api
