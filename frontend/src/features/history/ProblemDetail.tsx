import ProblemView from "@/components/ProblemView";
import api from "@/lib/api";
import type { Problem } from "@/types/Problems";
import axios from "axios";
import { useEffect, useState } from "react";
import { EmptyState } from "@/components/ui/empty-state"
import { useParams } from "react-router-dom";
import type { SubmissionHistory } from "@/types/Submissons";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Badge } from "@/components/ui/badge";
import { ChevronDown } from "lucide-react";
import DateToString from "@/util/dateToString";
import { LoadingSkeleton } from "@/components/ui/loading-skeleton"

async function getProblem(problemId: string): Promise<Problem | null> {
  try {
    const res = await api.get<Problem>(`/problems/${problemId}`);
    return res.data
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 404) {
      return null
    }
    throw err
  }
}

async function getSolutionsForProblem(problemId: string): Promise<SubmissionHistory | null> {
  try {
    const res = await api.get<SubmissionHistory>(`/submissions/${problemId}`);
    return res.data
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 404) {
      return null
    }
    throw err
  }
}


function ProblemDetail() {

  const { id } = useParams()
  const [problem, setProblem] = useState<Problem | null>()
  const [submissions, getSubmissions] = useState<SubmissionHistory | null>()

  const hasSubmissions = (submissions?.result.length ?? 0) > 0

  useEffect(() => {
    getProblem(id!).then(s => { setProblem(s) }).catch(() => setProblem(null))
    getSolutionsForProblem(id!).then(s => { getSubmissions(s) }).catch(() => getSubmissions(null))
  }, [id])

  if (problem === undefined) {
    return <LoadingSkeleton />
  }

  if (problem === null) {
    return (
      <EmptyState
        className="min-h-[70vh]"
        title="Said problem does not exist"
        description="Humm! Something is wrong this problem does not exist."
      />
    )
  }
  return (
    <div>
      <ProblemView problem={problem} isFromHistory={true} />
      {
        hasSubmissions ?
          <div className="mx-auto w-full max-w-5xl px-10 py-14 flex flex-col gap-1">
            <h2 className="font-semibold text-xl pb-8">Attempt History</h2>
            <div className="max-h-128 overflow-y-auto scrollbar-none flex flex-col gap-2">
              {
                submissions?.result.map((s, i) => (
                  <div>
                    <Collapsible
                      key={s.solution_id}
                      defaultOpen={i === 0}
                      className="rounded-lg border border-border-faint overflow-hidden"
                    >
                      <CollapsibleTrigger className="group flex w-full items-center justify-between gap-3 px-4 py-3 hover:bg-accent/50">
                        <div className="flex items-center gap-2.5">
                          <span className="w-4 font-mono text-xs text-text-faint">
                            #{submissions.result.length - i}
                          </span>
                          <Badge variant={s.status === "Solved" ? "solved" : "open"}>{s.status}</Badge>
                        </div>
                        <div className="flex items-center gap-3">
                          <span className="text-xs text-text-faint">{DateToString(s.submitted_at)}</span>
                          <ChevronDown className="size-4 text-text-faint transition-transform group-data-panel-open:rotate-180" />
                        </div>
                      </CollapsibleTrigger>
                      <CollapsibleContent className="border-t border-border-faint bg-code-surface">
                        <div className="flex">
                          <div className="shrink-0 bg-code-gutter px-2.5 py-3.5 text-right">
                            <div className="font-mono text-sm leading-[1.85] text-text-faint">
                            </div>
                          </div>
                          <pre className="flex-1 overflow-x-auto px-4 py-3.5 font-mono text-sm leading-[1.85] text-code-plain">
                            {s.solution}
                          </pre>
                        </div>
                      </CollapsibleContent>
                    </Collapsible>
                  </div>
                ))
              }
            </div>
          </div>
          : null
      }
    </div>
  )
}

export default ProblemDetail
