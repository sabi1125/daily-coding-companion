import type { AiHelp, Problem } from "@/types/Problems";
import { Badge } from "./ui/badge";
import ResolveBadgeVariant from "@/util/badgeResolver";
import { useEffect, useState } from "react";
import api, { getErrorMessage } from "@/lib/api";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
  SelectLabel,
} from "./ui/select"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "./ui/sheet";
import { LoadingHelp } from "@/components/ui/loading-help";
import { javascript } from "@codemirror/lang-javascript"
import { python } from "@codemirror/lang-python"
import { cpp } from "@codemirror/lang-cpp"
import { vim } from "@replit/codemirror-vim"
import CodeMirror, { type Extension } from "@uiw/react-codemirror"
import { codeEditorTheme } from "@/lib/codeMirrorTheme";
import { go } from "@codemirror/lang-go"
import { ToggleGroup } from "@/components/ui/toggle-group"
import { ToggleGroupItem } from "@/components/ui/toggle-group";
import type { SubmissionRequest, ExecuteSubmittedSolution, ExecuteResposne } from "@/types/Submissons";
import { useNavigate } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";

type LanguageType = "javascript" | "python" | "cpp" | "go"

async function getHelp(problemId: string): Promise<AiHelp> {
  const res = await api.get<AiHelp>(`/problems/${problemId}/help`)
  return res.data
}

async function postSolution(problemId: string, solution: string, status: string) {
  const req: SubmissionRequest = {
    solution: solution,
    status: status
  }
  await api.post(`/submissions/${problemId}`, req)
}

async function runSolution(language: string, solution: string): Promise<ExecuteResposne> {
  const req: ExecuteSubmittedSolution = {
    language: language,
    content: solution
  }
  const res = await api.post("/submissions/run", req)
  return res.data
}

const languages = [
  { label: "python", value: "python" as LanguageType },
  { label: "javascript", value: "javascript" as LanguageType },
  { label: "cpp", value: "cpp" as LanguageType },
  { label: "go", value: "go" as LanguageType }
]

const language: Record<LanguageType, Extension> = {
  javascript: javascript(),
  python: python(),
  cpp: cpp(),
  go: go(),
}

function ProblemView({ problem, isFromHistory }: { problem: Problem, isFromHistory: boolean }) {
  const [help, setHelp] = useState<AiHelp | null>()
  const [helpOpen, setHelpOpen] = useState(false)
  const [code, setCode] = useState<LanguageType>("python")
  const [solution, setSolution] = useState<string>()
  const [status, setStatus] = useState<string>("Failed")
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [output, setOutput] = useState<ExecuteResposne>()
  const [isRunning, setIsRunning] = useState(false)
  const [showOutput, setShowOutput] = useState(false)

  const navigate = useNavigate()

  const handleRun = () => {
    if (!solution || isRunning) return
    setIsRunning(true)
    setShowOutput(true)
    runSolution(code, solution)
      .then(setOutput)
      .catch((err) => {
        toast.error(getErrorMessage(err, "Couldn't run code. Please try again later."))
        setShowOutput(false)
      })
      .finally(() => setIsRunning(false))
  }

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
        e.preventDefault()
        handleRun()
      }
    }
    window.addEventListener("keydown", handler)
    return () => window.removeEventListener("keydown", handler)
  })

  return (
    <div className="mx-auto w-full max-w-5xl px-10 py-14 flex flex-col gap-5">

      {
        isFromHistory ?
          <span onClick={() => { navigate(-1) }} className="text-text-faint text-xs font-semibold hover:text-foreground cursor-pointer">Back to history</span>
          : null
      }
      {/* Badge */}

      <section className="flex flex-row justify-between items-center">
        <div className="flex flex-row gap-1">
          <Badge variant={"review"}>{problem.difficulty}</Badge>
          <Badge variant={"disabled"}>{problem.algorithm_tag}</Badge>
          {problem.needs_review_flag ?
            <Badge variant={"review"}>Need Review</Badge>
            : null}
        </div>
        <div className="flex flex-row gap-2">
          <Badge variant={ResolveBadgeVariant(problem.status)}>
            {problem.status}
          </Badge>
          <Sheet
            open={helpOpen}
            onOpenChange={(open) => {
              setHelpOpen(open)
              if (open) {
                setHelp(null);
                getHelp(problem.problem_id).then(v => setHelp(v))
                  .catch((err) => { toast.error(getErrorMessage(err, "Couldn't get AI help. Please try again later.")); setHelp(null) })
              }
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
                {help === null && <LoadingHelp />}
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

      <h2 className="text-2xl font-semibold text-foreground">{problem.title}</h2>
      <p className="text-base text-muted-foreground whitespace-pre-line">{problem.problem_text}</p>
      <h3 className="font-semibold text-sm">Submit an attempt</h3>

      {/* editor */}

      <div className="rounded-lg border border-border-faint overflow-hidden shadow-2xs">
        <div className="flex flex-row justify-between items-center text-text-secondary h-11 p-4 border-b border-border-faint">
          <div>
            <span className="text-sm">solution</span>
          </div>
          <div className="flex flex-row gap-3">
            <Select value={code} onValueChange={(v => setCode(v ? v : "python"))}>
              <SelectTrigger className="w-full max-w-48">
                <SelectValue className="font-semibold text-xs" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectLabel className="font-semibold text-xs">{code}</SelectLabel>
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
            <Button
              variant="outline"
              className="text-sm w-20"
              disabled={!solution || isRunning}
              onClick={handleRun}
            >
              {isRunning ? <Loader2 className="size-4 animate-spin" /> : <>Run <p className="text-text-faint text-xs">⌘↵</p></>}
            </Button>
          </div>
        </div>
        <div>
          <CodeMirror
            height="450px"
            extensions={[codeEditorTheme, vim(), language[code]]}
            onChange={s => { setSolution(s) }}
            basicSetup={{
              lineNumbers: true,
            }}
          />
          {
            showOutput ?
              <div className="h-30 bg-foreground text-foreground flex flex-col overflow-x-auto scrollbar-none pr-4 pl-4 pt-3">
                {/* header */}
                <div className="text-muted-foreground font-semibold flex flex-row gap-2 justify-between items-center text-xs">
                  <div>
                    <p className="text-muted-foreground">output</p>
                  </div>
                  <div className="flex flex-row items-center justify-center gap-2">
                    <p>{output?.run.cpu_time}ms</p>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="cursor-pointer"
                      onClick={() => { setShowOutput(false) }}
                    >x</Button>
                  </div>
                </div>
                <div className="text-background pr-2">
                  <p className="text-xs whitespace-pre-line">{output?.run.output}</p>
                </div>
              </div>
              : ""
          }
        </div>
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
          disabled={isSubmitting}
          onClick={
            () => {
              setIsSubmitting(true)
              postSolution(problem.problem_id, solution!, status).then(() => toast.success("Submitted successfully!")).catch((err) => { toast.error(getErrorMessage(err, "Couldn't submit solution. Please try again later.")); }).finally(() => setIsSubmitting(false))
            }
          }
          className="py-4 text-xs">
          {isSubmitting ? <Loader2 className="size-4 animate-spin" /> : "Submit attempt"}
        </Button>
      </div>
    </div >
  )
}

export default ProblemView
