export interface TodaysProblem {
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
