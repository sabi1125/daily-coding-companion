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

export function getErrorMessage(err: unknown, fallback: string): string {
  if (axios.isAxiosError(err) && typeof err.response?.data?.message === "string") {
    return err.response.data.message
  }
  return fallback
}

export default api
