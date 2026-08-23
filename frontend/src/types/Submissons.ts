export interface SubmissionRequest {
  solution: string
  status: string
}

export interface Submission {
  solution_id: string
  problem_id: string
  solution: string
  status: string
  submitted_at: string
}

export interface SubmissionHistory {
  result: Submission[]
  total: number
}
