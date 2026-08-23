export const Status = {
  Open: "Open",
  Failed: "Failed",
  Solved: "Solved",
  All: "All"
} as const

export type StatusType = typeof Status[keyof typeof Status]

export interface ProblemHistory {
  problem_id: string,
  title: string | null,
  status: StatusType,
  needs_review_flag: boolean,
  created_at: string
}

export interface ProblemResponse {
  result: ProblemHistory[],
  total: number
}

export interface Problem {
  problem_id: string,
  raw_problem: string,
  title: string | null,
  problem_text: string,
  algorithm_tag: string,
  difficulty: string,
  status: string,
  ai_help: string,
  needs_review_flag: boolean,
  created_at: string,
  updated_at: string
}

export interface AiHelp {
  concept: string,
  nudge: string,
  approach: string,
  walkthrough: string
}
