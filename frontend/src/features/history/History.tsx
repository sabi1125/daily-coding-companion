import { ToggleGroupItem } from "@/components/ui/toggle-group"
import { ToggleGroup } from "@/components/ui/toggle-group"
import { type ProblemResponse } from "@/types/Problems"
import { useEffect, useState } from "react"
import api, { getErrorMessage } from "@/lib/api"
import { EmptyState } from "@/components/ui/empty-state"
import { Badge } from "@/components/ui/badge"
import ResolveBadgeVariant from "@/util/badgeResolver"
import { useNavigate } from "react-router-dom"
import DateToString from "@/util/dateToString"
import { LoadingHistory } from "@/components/ui/loading-history"
import { toast } from "sonner"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Button } from "@/components/ui/button"
import { IoMdClose } from 'react-icons/io';
import { FaCaretDown } from 'react-icons/fa';

function createParamString(difficulty: string[]): string {
  let paramString = ""
  for (const value of difficulty) {
    paramString = paramString + "&difficulty=" + value
  }
  return paramString
}

async function getUserProblems(status: string, difficulty: string[]): Promise<ProblemResponse> {
  if (status == "All") {
    status = ""
  }

  if (difficulty.length > 0) {
    const param = createParamString(difficulty)
    const res = await api.get(`/problems?status=${status}${param}`)
    return res.data
  } else {
    const res = await api.get(`/problems?status=${status}`)
    return res.data
  }
}

function History() {

  const navigate = useNavigate()

  const [problems, setProblems] = useState<ProblemResponse | null>()
  const [status, setStatus] = useState<string>("All")
  const [difficulty, setDifficulty] = useState<string[]>([])
  const difficultyList = ["Easy", "Medium", "Hard"]

  const tabStates = ["All", "Open", "Failed", "Solved"]

  const editDifficultyList = async (diffString: string) => {
    const index = difficulty.indexOf(diffString)
    if (index !== -1) {
      setDifficulty(prevList => prevList.filter(item => item !== diffString));
    } else {
      setDifficulty(prevList => [...prevList, diffString])
    }
  }

  useEffect(() => {
    getUserProblems(status, difficulty)
      .then(s => { setProblems(s) })
      .catch((err) => { toast.error(getErrorMessage(err, "Couldn't get problems history. Please try again later.")); setProblems(null) })
  }, [status, difficulty])

  if (problems === undefined) {
    return <LoadingHistory />
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

      <div className="flex flex-row justify-between">
        <div>
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
          {
            difficulty.length === 0 ?
              null
              :
              <div className="pt-5">
                {
                  difficulty.map((v) => (
                    <Button onClick={() => {
                      editDifficultyList(v)
                    }}
                      size="xs"
                    >{v}
                      <IoMdClose />
                    </Button>
                  ))
                }
              </div>
          }
        </div>
        <div>
          <Popover>
            <PopoverTrigger render={<Button variant="outline" size="xs">Difficulty <FaCaretDown /></Button>} />
            <PopoverContent className="w-44 p-1.5">
              <div className="flex flex-row justify-around">
                {
                  difficultyList.map(
                    v => (<Button variant="outline" size="xs" onClick={() => { editDifficultyList(v); getUserProblems(status, difficulty) }}>{v}</Button>))
                }
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>

      { /* history table */}

      <section className="max-h-128 overflow-y-auto scrollbar-none flex flex-col">
        {hasProblems ? problems.result.map(p => (
          // when problems exist
          <section
            key={p.problem_id}
            className="border-b border-border-faint py-5 flex flex-row items-center justify-between gap-4 cursor-pointer"
            onClick={() => navigate(`/history/${p.problem_id}`)}
          >
            <div className="flex flex-row rounded-lg items-center gap-3 min-w-0">
              <h2 className="font-medium">{p.title}</h2>
              {p.needs_review_flag ? <Badge variant={"review"}>Review</Badge> : null}
            </div>
            <div className="flex flex-row gap-2 shrink-0 items-center">
              <p className="text-xs text-text-faint whitespace-nowrap">{DateToString(p.created_at)}</p>
              <Badge variant={ResolveBadgeVariant(p.status)}>{p.status}</Badge>
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
