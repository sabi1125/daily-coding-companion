import LoginState from "@/components/loginState"
import { LoadingSkeleton } from "@/components/ui/loading-skeleton"
import api, { getErrorMessage } from "@/lib/api"
import Router from "@/router/Router"
import type { SettingResponse } from "@/types/SettingResponse"
import axios from "axios"
import { useEffect, useState } from "react"
import { toast } from "sonner"

async function getSetting(): Promise<SettingResponse | null> {
  try {
    const res = await api.get<SettingResponse>("/settings")
    return res.data
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 401) {
      return null
    }
    throw err
  }
}

function Auth() {
  const [setting, setSetting] = useState<SettingResponse | null>()

  useEffect(() => {
    let ignore = false
    getSetting()
      .then(s => { if (!ignore) setSetting(s) })
      .catch((err) => { if (!ignore) { toast.error(getErrorMessage(err, "Couldn't reach the server. Please try again later.")); setSetting(null) } })
    return () => { ignore = true }
  }, [])

  if (setting === undefined) {
    return <LoadingSkeleton />
  }

  if (setting === null) {
    return (
      <LoginState />
    )
  }

  if (setting.needs_reauth) {
    return (
      <LoginState />
    )
  }

  return <Router />
}

export default Auth
