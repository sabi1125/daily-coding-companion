import { Skeleton } from "@/components/ui/skeleton"

const sections = ["Concept", "Nudge", "Approach", "Walkthrough"]

export function LoadingHelp() {
  return (
    <div className="flex flex-col gap-1 px-15 pb-4">
      {sections.map((section, i) => (
        <div
          key={section}
          className={`flex flex-col gap-3 py-6 ${i < sections.length - 1 ? "border-b border-border-faint" : ""}`}
        >
          <Skeleton className="h-5 w-24" />
          <div className="flex flex-col gap-2">
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-2/3" />
          </div>
        </div>
      ))}
    </div>
  )
}
