import { ToggleGroupItem } from "@/components/ui/toggle-group"
import { ToggleGroup } from "@/components/ui/toggle-group"
import { type ProblemResponse } from "@/types/ProblemsResponse"
import { useEffect, useState } from "react"
import api from "@/lib/api"
import { Badge, badgeVariants } from "@/components/ui/badge"
import type { VariantProps } from "class-variance-authority"
import { EmptyState } from "@/components/ui/empty-state"

async function getUserProblems(status: string): Promise<ProblemResponse> {
  if (status == "All") {
    status = ""
  }
  const res = await api.get(`/problems?status=${status}`)
  return res.data
}

function dateToString(date: string): string {
  const dateObj = new Date(date)
  const formattedDate = dateObj.toLocaleDateString("en-US", {
    month: "short", // "Aug"
    day: "numeric", // "21"
    year: "numeric", // "2026"
  });

  return formattedDate
}

function resolveBadgeVariant(status: string): VariantProps<typeof badgeVariants>["variant"] {
  if (status === "Solved") return "solved"
  if (status === "Failed") return "failed"
  return "open"
}

function History() {

  const [problems, setProblems] = useState<ProblemResponse | null>()
  const [status, setStatus] = useState<string>("All")

  const tabStates = ["All", "Open", "Failed", "Solved"]

  useEffect(() => {
    getUserProblems(status).then(s => {
      setProblems(s);
    })
  }, [status])

  const hasProblems = (problems?.result.length ?? 0) > 0

  return (
    <div className="mx-auto w-full max-w-3xl px-10 py-14 flex flex-col gap-8">
      {/* Header */}
      <h1 className="text-2xl font-semibold">History</h1>

      {/* Toggle bar */}

      {(status === "All" && !hasProblems) ? null :
        <ToggleGroup className={"rounded-lg bg-secondary p-1 gap-1"} value={[status]} onValueChange={(v) => {
          if (v.length === 0) return
          setStatus(v[0] ?? "All")
        }}>
          {tabStates.map(t => (
            <ToggleGroupItem value={t}
              key={t}
              variant={"outline"}
              className="h-6 px-3 text-xs rounded-md text-muted-background bg-background
           data-pressed:bg-foreground data-pressed:text-background data-pressed:shadow-sm
           not-data-pressed:hover:bg-transparent">
              {t}
            </ToggleGroupItem>
          ))}
        </ToggleGroup>
      }

      { /* history table */}

      <section className="max-h-128 overflow-y-auto scrollbar-none flex flex-col">
        {hasProblems ? problems!.result.map(p => (
          // when problems exist
          <section key={p.problem_id} className="border-b border-border-faint py-5 flex flex-row justify-between" >
            <div className="flex flex-row rounded-lg items-center gap-3">
              <h2 className="font-medium">{p.title}</h2>
              {p.needs_review_flag ? <Badge variant={"review"}>Review</Badge> : null}
            </div>
            <div className="flex flex-row gap-2">
              <p className="text-xs text-text-faint">{dateToString(p.created_at)}</p>
              <Badge variant={resolveBadgeVariant(p.status)} className="">{p.status}</Badge>
            </div>
          </section>
        )) :
          <EmptyState
            className="w-full h-100"
            title="No problems yet"
            description="Once your first daily problem is ingested, it'll show up here."
          />
        }
      </section>
    </div >
  )
}

export default History
