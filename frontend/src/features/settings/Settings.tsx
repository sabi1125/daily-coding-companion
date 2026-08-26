import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import type { SettingResponse, PatchSettingRequest } from "@/types/SettingResponse"
import api, { getErrorMessage } from "@/lib/api"
import { env } from "@/lib/env"
import { useEffect, useState } from "react"
import { EmptyState } from "@/components/ui/empty-state"
import { LoadingSetting } from "@/components/ui/loading-setting"
import { Loader2 } from "lucide-react"
import { toast } from "sonner"

async function getSetting(): Promise<SettingResponse> {
  const res = await api.get<SettingResponse>("/settings");
  return res.data
}

async function patchSetting(pref: string) {
  const req: PatchSettingRequest = { get_help_preferences: pref }
  await api.patch("/settings", req)
}


function Settings() {
  const [setting, setSettings] = useState<SettingResponse | null | undefined>(undefined)
  const [pref, setPref] = useState<string>()
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    getSetting()
      .then(s => { setSettings(s); setPref(s.get_help_preferences ?? "") })
      .catch((err) => { toast.error(getErrorMessage(err, "Couldn't load settings. Please try again later.")); setSettings(null) })
  }, [])

  if (setting === undefined) {
    return <LoadingSetting />
  }

  if (setting === null) {
    return (
      <EmptyState
        className="min-h-[70vh]"
        title="Couldn't load settings"
        description="Something went wrong loading your settings. Try refreshing the page."
      />
    )
  }


  return (
    <div className="mx-auto w-full max-w-2xl px-10 py-14 flex flex-col gap-8">
      {/* Header */}

      <h1 className="text-2xl font-semibold">Settings</h1>

      {/* Active */}

      <section className="gap-2 flex flex-col">
        <h2 className="text-sm font-medium text-foreground">Account</h2>
        <div className="rounded-lg border border-border-faint p-4 flex justify-between items-center gap-2">
          <p className="text-sm">{setting.email}</p>
          {setting.needs_reauth ?
            <Button variant="default"
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
              onClick={() => api.post("/auth/signout")
                .then(() => window.location.assign("/"))
                .catch((err) => toast.error(getErrorMessage(err, "Error occurred while signing out. Please try again later.")))
              }
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
      </section >

      {/* Daily Ingest */}

      < section className="gap-2 flex flex-col" >
        <h2 className="text-sm font-medium text-foreground">Daily Ingest</h2>
        <div className="rounded-lg border border-border-faint p-4 flex justify-between items-center gap-4">
          <div className="flex flex-col gap-1">
            <p className="text-sm">Runs automatically every day</p>
            <p className="text-xs text-muted-foreground">Nothing to turn on or off — a new problem is ingested daily as long as you're connected</p>
          </div>
          <Badge variant="disabled" className="shrink-0">Automatic</Badge>
        </div>
      </section >

      {/* Stats */}

      < section className="flex flex-col gap-2" >
        <h2 className="text-sm font-medium text-foreground">Stats</h2>
        <div className="grid grid-cols-2 gap-3">
          <div className="rounded-lg border border-border-faint p-4 flex flex-col gap-1">
            <p className="text-3xl font-semibold">{setting.solved_count}</p>
            <p className="text-sm text-muted-foreground">Solved</p>
          </div>
          <div className="rounded-lg border border-border-faint p-4 flex flex-col gap-1">
            <p className="text-3xl font-semibold">{setting.unsolved_count}</p>
            <p className="text-sm text-muted-foreground">Unsolved</p>
          </div>
        </div>
      </section >

      {/* Preferences */}

      < section className="gap-2 flex flex-col" >
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
          placeholder={setting.get_help_preferences ? setting.get_help_preferences : "e.g. Prefer Python examples. Keep explanations brief."} />
        <Button variant="default"
          disabled={isSaving}
          onClick={() => {
            setIsSaving(true)
            patchSetting(pref ?? "")
              .then(() => toast.success("Preferences successfully saved."))
              .catch((err) => toast.error(getErrorMessage(err, "Something went wrong when saving preferences. Please try again later.")))
              .finally(() => {
                setIsSaving(false)
              })
          }}
          className="
        text-primary-foreground text-sm
        px-3
        py-4
        w-20
        h-8
        font-medium
            hover:bg-muted hover:text-primary">
          {isSaving ? <Loader2 className="size-4 animate-spin" /> : "Save"}
        </Button>
      </section >
    </div >
  )
}

export default Settings
