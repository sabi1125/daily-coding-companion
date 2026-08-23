import { Badge } from "@/components/ui/badge";
import api from "@/lib/api";
import axios from "axios";
import { type AiHelp, type TodaysProblem } from "@/types/TodaysProblem"
import { Button } from "@/components/ui/button";
import { useEffect, useState } from "react"
import { javascript } from "@codemirror/lang-javascript"
import { python } from "@codemirror/lang-python"
import { cpp } from "@codemirror/lang-cpp"
import { vim } from "@replit/codemirror-vim"
import CodeMirror, { type Extension } from "@uiw/react-codemirror"
import { codeEditorTheme } from "@/lib/codeMirrorTheme";
import { ToggleGroup } from "@/components/ui/toggle-group"
import { ToggleGroupItem } from "@/components/ui/toggle-group";
import { go } from "@codemirror/lang-go"
import type { Submission } from "@/types/Submissons"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { EmptyState } from "@/components/ui/empty-state";

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
  SelectLabel,
} from "@/components/ui/select"

type LanguageType = "javascript" | "python" | "cpp" | "go"

async function getTodaysProblem(): Promise<TodaysProblem | null> {
  try {
    const res = await api.get<TodaysProblem>("/problems/today");
    return res.data
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 404) {
      return null
    }
    throw err
  }
}

async function postSolution(problemId: string, solution: string, status: string) {
  const req: Submission = {
    solution: solution,
    status: status
  }
  await api.post(`/submissions/${problemId}`, req)
}

async function getHelp(problemId: string): Promise<AiHelp> {
  const res = await api.get<AiHelp>(`/problems/${problemId}/help`)
  return res.data
}

const language: Record<LanguageType, Extension> = {
  javascript: javascript(),
  python: python(),
  cpp: cpp(),
  go: go(),
}

function Today() {
  const [problem, setProblem] = useState<TodaysProblem | null>()
  const [code, setCode] = useState<LanguageType>("python")
  const [solution, setSolution] = useState<string>()
  const [status, setStatus] = useState<string>("Failed")
  const [help, setHelp] = useState<AiHelp | null>()
  const [helpOpen, setHelpOpen] = useState(false)

  useEffect(() => {
    getTodaysProblem().then(s => { setProblem(s) }).catch(() => setProblem(null))
  }, [])


  const languages = [
    { label: "python", value: "python" as LanguageType },
    { label: "javascript", value: "javascript" as LanguageType },
    { label: "cpp", value: "cpp" as LanguageType },
    { label: "go", value: "go" as LanguageType }
  ]


  return (
    <div className="mx-auto w-full max-w-5xl px-10 py-14 flex flex-col gap-5">

      {problem === null && (
        <EmptyState
          className="min-h-[70vh]"
          title="No problem today"
          description="This refreshes automatically — check back tomorrow."
        />
      )}

      {problem !== null && (
        <>
          {/* badges */}

          <section className="flex flex-row justify-between items-center">
            <div className="flex flex-row gap-1">
              <Badge variant={"review"}>{problem?.difficulty}</Badge>
              <Badge variant={"disabled"}>{problem?.algorithm_tag}</Badge>
              {problem?.needs_review_flag ?
                <Badge variant={"review"}>Need Review</Badge>
                : null}
            </div>
            <div className="flex flex-row gap-2">
              <Badge variant={"open"}>
                {problem?.status}
              </Badge>
              <Sheet
                open={helpOpen}
                onOpenChange={(open) => {
                  setHelpOpen(open)
                  if (open) { setHelp(null); getHelp(problem!.problem_id).then(v => setHelp(v)) }
                }}
              >
                <SheetTrigger render={<Button size={"xxs"} variant={"outline"} />}>
                  Get help
                </SheetTrigger>
                <SheetContent className="data-[side=right]:xl:max-w-5xl gap-1">
                  <SheetHeader className="pb-0 pt-20">
                    <SheetTitle className="px-10 text-2xl font-semibold">Get Help</SheetTitle>
                  </SheetHeader>
                  <div className="flex flex-col gap-1 px-15 pb-4 overflow-y-auto">
                    {help && (
                      <>
                        <section className="flex flex-col gap-3 border-b border-border-faint py-6">
                          <h2 className="text-xl font-medium text-foreground">Concept</h2>
                          <p className="text-sm text-muted-foreground whitespace-pre-line">{help.concept}</p>
                        </section>
                        <section className="flex flex-col gap-3 border-b border-border-faint py-6">
                          <h2 className="text-xl font-medium text-foreground">Nudge</h2>
                          <p className="text-sm text-muted-foreground whitespace-pre-line">{help.nudge}</p>
                        </section>
                        <section className="flex flex-col gap-3 border-b border-border-faint py-6">
                          <h2 className="text-xl font-medium text-foreground">Approach</h2>
                          <p className="text-sm text-muted-foreground whitespace-pre-line">{help.approach}</p>
                        </section>
                        <section className="flex flex-col gap-3 py-6">
                          <h2 className="text-xl font-medium text-foreground">Walkthrough</h2>
                          <p className="text-sm text-muted-foreground whitespace-pre-line">{help.walkthrough}</p>
                        </section>
                      </>
                    )}
                  </div>
                </SheetContent>
              </Sheet>
            </div>
          </section>

          {/* problem details */}

          <h2 className="text-2xl font-semibold text-foreground">{problem?.title}</h2>
          <p className="text-base text-muted-foreground">{problem?.problem_text}</p>
          <h3 className="font-semibold text-sm">Submit an attempt</h3>

          {/* editor */}

          <div className="rounded-lg border border-border-faint overflow-hidden shadow-2xs">
            <div className="flex flex-row justify-between items-center text-text-secondary h-9 p-4 border-b border-border-faint">
              <span className="text-sm">solution.py</span>

              <Select value={code} onValueChange={(v => setCode(v!))}>
                <SelectTrigger className="w-full max-w-48">
                  <SelectValue className="font-semibold text-xs" />
                </SelectTrigger>
                <SelectContent>
                  <SelectGroup>
                    <SelectLabel className="font-semibold text-xs">Python</SelectLabel>
                    {languages.map((lang) => (
                      <SelectItem
                        key={lang.label}
                        value={lang.label}
                        className="font-semibold text-xs">
                        {lang.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
            <CodeMirror
              height="450px"
              extensions={[codeEditorTheme, vim(), language[code]]}
              onChange={s => { setSolution(s) }}
              basicSetup={{
                lineNumbers: true,
              }}
            />
          </div>

          {/* toggle group */}
          <div className="flex flex-row justify-between items-center">
            <ToggleGroup
              className="rounded-lg bg-secondary p-1 gap-1"
              value={[status]}
              onValueChange={v => { if (v.length === 0) return; setStatus(v[0]) }}
            >
              <ToggleGroupItem
                className="rounded-md h-7 px-3 text-xs text-muted-foreground data-pressed:bg-foreground data-pressed:text-background data-pressed:shadow-sm"
                value="Failed"
              >
                Not Solved
              </ToggleGroupItem>
              <ToggleGroupItem
                className="rounded-md h-7 px-3 text-xs text-muted-foreground data-pressed:bg-foreground data-pressed:text-background data-pressed:shadow-sm"
                value="Solved"
              >Solved</ToggleGroupItem>
            </ToggleGroup>
            <Button
              onClick={
                () => {
                  postSolution(problem!.problem_id, solution!, status)
                }
              }
              className="py-4 text-xs">Submit attempt</Button>
          </div>
        </>
      )}
    </div >
  )
}
export default Today
