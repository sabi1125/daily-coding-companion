import ProblemView from "@/components/ProblemView";
import api from "@/lib/api";
import type { Problem } from "@/types/Problems";
import axios from "axios";
import { useEffect, useState } from "react";
import { EmptyState } from "@/components/ui/empty-state"
import { useParams } from "react-router-dom";

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


function ProblemDetail() {

  const { id } = useParams()
  const [problem, setProblem] = useState<Problem | null>()

  useEffect(() => {
    getProblem(id!).then(s => { setProblem(s) }).catch(() => setProblem(null))
  }, [id])

  if (problem === undefined) {
    return <p>Loading...</p>
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
    <ProblemView problem={problem} isFromHistory={true} />
  )
}

export default ProblemDetail
