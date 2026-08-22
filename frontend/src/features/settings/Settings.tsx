import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import type { SettingResponse, PatchSettingRequest } from "@/types/SettingResponse"
import api from "@/lib/api"
import { env } from "@/lib/env"
import { useEffect, useState } from "react"

// TODO: needs loading spinners after implemented

async function getSetting(): Promise<SettingResponse> {
  // TODO: Needs error handling after implemented
  const res = await api.get<SettingResponse>("/settings");
  return res.data;
}

async function patchSetting(pref: string) {
  // TODO: Needs error handling after implemented
  const req: PatchSettingRequest = { get_help_preferences: pref }
  await api.patch("/settings", req)
}


function Settings() {
  const [setting, setSettings] = useState<SettingResponse | null>(null)
  const [pref, setPref] = useState<string>()

  useEffect(() => {
    getSetting().then(s => { setSettings(s); setPref(s.get_help_preferences ?? "") })
  }, [])

  return (
    <div className="mx-auto w-full max-w-2xl px-10 py-14 flex flex-col gap-8">
      {/* Header */}

      <h1 className="text-2xl font-semibold">Settings</h1>

      {/* Active */}

      <section className="gap-2 flex flex-col">
        <h2 className="text-sm font-medium text-foreground">Account</h2>
        <div className="rounded-lg border border-border-faint p-4 flex justify-between items-center gap-2">
          <p className="text-sm">{setting?.email}</p>
          {setting?.needs_reauth ?
            <Button variant="default"
              // TODO: need to find better way to use window.location
              // REFAC: this needs refactoring
              onClick={() => window.location.assign(`${env.apiBaseUrl}/auth/google`)}
              className="
            text-primary-foreground text-sm 
            px-3 h-8 
            font-medium 
            hover:bg-muted hover:text-primary">
              Reconnect
            </Button>
            :
            <Button variant="outline"
              // TODO: need to redirect to actual login page after it is created
              // REFAC: this needs refactoring
              onClick={() => api.post("/auth/signout").then(() => window.location.assign("/"))}
              className="
            text-destructive text-sm 
            border-destructive/30 rounded-md 
            px-3 h-8 
            font-medium 
            hover:bg-destructive/5 hover:text-destructive">
              Sign out
            </Button>
          }
        </div>

        <p className="text-xs text-muted-foreground">Coding Companion reads your daily problem email — a subscription is required.</p>
      </section>

      {/* Daily Ingest */}

      <section className="gap-2 flex flex-col">
        <h2 className="text-sm font-medium text-foreground">Daily Ingest</h2>
        <div className="rounded-lg border border-border-faint p-4 flex justify-between items-center gap-4">
          <div className="flex flex-col gap-1">
            <p className="text-sm">Runs automatically every day</p>
            <p className="text-xs text-muted-foreground">Nothing to turn on or off — a new problem is ingested daily as long as you're connected</p>
          </div>
          <Badge variant="disabled" className="shrink-0">Automatic</Badge>
        </div>
      </section>

      {/* Stats */}

      <section className="flex flex-col gap-2">
        <h2 className="text-sm font-medium text-foreground">Stats</h2>
        <div className="grid grid-cols-2 gap-3">
          <div className="rounded-lg border border-border-faint p-4 flex flex-col gap-1">
            <p className="text-3xl font-semibold">{setting?.solved_count}</p>
            <p className="text-sm text-muted-foreground">Solved</p>
          </div>
          <div className="rounded-lg border border-border-faint p-4 flex flex-col gap-1">
            <p className="text-3xl font-semibold">{setting?.unsolved_count}</p>
            <p className="text-sm text-muted-foreground">Unsolved</p>
          </div>
        </div>
      </section>

      {/* Preferences */}

      <section className="gap-2 flex flex-col">
        <h2 className="text-sm font-medium text-foreground">Get Help preferences</h2>
        <p className="text-xs text-muted-foreground">
          Optional — appended to every future Get Help request on top of the required concept explanation. It extends the response, it doesn't replace anything.
        </p>
        <textarea
          rows={3}
          value={pref}
          onChange={(e) => setPref(e.target.value)}
          className="
            border border-border-faint rounded-lg
            w-full p-3
            text-sm placeholder:text-text-faint
            focus-visible:outline-none
            resize-none"
          placeholder={setting?.get_help_preferences ? setting.get_help_preferences : "e.g. Prefer Python examples. Keep explanations brief."} />
        <Button variant="default"
          onClick={() => patchSetting(pref ?? "")}
          className="
        text-primary-foreground text-sm
        px-3
        py-4
        w-20
        h-8
        font-medium 
            hover:bg-muted hover:text-primary">
          Save
        </Button>
      </section>
    </div >
  )
}

export default Settings
