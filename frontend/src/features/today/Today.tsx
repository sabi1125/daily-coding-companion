import api from "@/lib/api";
import axios from "axios";
import { type Problem } from "@/types/Problems"
import { useEffect, useState } from "react"
import { EmptyState } from "@/components/ui/empty-state"
import ProblemView from "@/components/ProblemView";

async function getTodaysProblem(): Promise<Problem | null> {
  try {
    const res = await api.get<Problem>("/problems/today");
    return res.data
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 404) {
      return null
    }
    throw err
  }
}


function Today() {
  const [problem, setProblem] = useState<Problem | null>()

  useEffect(() => {
    getTodaysProblem().then(s => { setProblem(s) }).catch(() => setProblem(null))
  }, [])

  if (problem === undefined) {
    return <p>Loading...</p>
  }

  if (problem === null) {
    return (
      <EmptyState
        className="min-h-[70vh]"
        title="No problem today"
        description="This refreshes automatically — check back tomorrow."
      />
    )
  }

  return (
    <ProblemView problem={problem} isFromHistory={false} />
  )
}
export default Today
