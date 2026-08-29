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

export interface ExecuteSubmittedSolution {
  language: string
  content: string
}

export interface ExecuteResposne {
  language: string
  version: string
  run: Run
}

export interface Run {
  stdout: string
  stderr: string
  output: string
  code: number
  signal: string | null
  cpu_time: number
  wall_time: number
  memory: number
  message: string | null
  status: string | null
}
