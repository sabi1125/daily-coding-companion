import { ToggleGroupItem } from "@/components/ui/toggle-group"
import { ToggleGroup } from "@/components/ui/toggle-group"
import { type ProblemResponse } from "@/types/Problems"
import { useEffect, useState } from "react"
import api from "@/lib/api"
import { EmptyState } from "@/components/ui/empty-state"
import { Badge } from "@/components/ui/badge"
import ResolveBadgeVariant from "@/util/badgeResolver"
import { useNavigate } from "react-router-dom"
import DateToString from "@/util/dateToString"

async function getUserProblems(status: string): Promise<ProblemResponse> {
  if (status == "All") {
    status = ""
  }
  const res = await api.get(`/problems?status=${status}`)
  return res.data
}

function History() {

  const navigate = useNavigate()

  const [problems, setProblems] = useState<ProblemResponse | null>()
  const [status, setStatus] = useState<string>("All")

  const tabStates = ["All", "Open", "Failed", "Solved"]

  useEffect(() => {
    getUserProblems(status).then(s => {
      setProblems(s);
    })
  }, [status])

  if (problems === undefined) {
    return <p>Loading..</p>
  }

  if (problems === null) {
    return (
      <EmptyState
        className="min-h-[70vh]"
        title="No problems here"
        description="There is no history. Once a problem has been ingested the problem's history will be shown here."
      />
    )
  }

  const hasProblems = (problems.result.length ?? 0) > 0

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
        {hasProblems ? problems.result.map(p => (
          // when problems exist
          <section
            key={p.problem_id}
            className="border-b border-border-faint py-5 flex flex-row justify-between cursor-pointer"
            onClick={() => navigate(`/history/${p.problem_id}`)}
          >
            <div className="flex flex-row rounded-lg items-center gap-3">
              <h2 className="font-medium">{p.title}</h2>
              {p.needs_review_flag ? <Badge variant={"review"}>Review</Badge> : null}
            </div>
            <div className="flex flex-row gap-2">
              <p className="text-xs text-text-faint">{DateToString(p.created_at)}</p>
              <Badge variant={ResolveBadgeVariant(p.status)} className="">{p.status}</Badge>
            </div>
          </section>
        )) :
          <EmptyState
            className="w-full h-100"
            title={`No ${status === "All" ? "" : status} problem yet`}
            description={`Once there is a ${status === "All" ? "" : status} problem  it will show up here.`}
          />
        }
      </section>
    </div >
  )
}

export default History
