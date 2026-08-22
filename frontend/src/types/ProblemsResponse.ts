export const Status = {
  Open: "Open",
  Failed: "Failed",
  Solved: "Solved",
  All: "All"
} as const

export type StatusType = typeof Status[keyof typeof Status]

export interface Problem {
  problem_id: string,
  title: string,
  status: StatusType,
  needs_review_flag: boolean,
  created_at: string
}

export interface ProblemResponse {
  result: Problem[],
  total: number
}
