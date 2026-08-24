import { Skeleton } from "@/components/ui/skeleton"

export function LoadingHistory() {
  return (
    <div className="mx-auto w-full max-w-3xl px-10 py-14 flex flex-col gap-8">
      {/* Header */}

      <Skeleton className="h-8 w-28" />

      {/* Toggle bar */}

      <div className="rounded-lg bg-secondary p-1 gap-1 flex w-fit">
        <Skeleton className="h-6 w-10" />
        <Skeleton className="h-6 w-12" />
        <Skeleton className="h-6 w-12" />
        <Skeleton className="h-6 w-14" />
      </div>

      {/* History list */}

      <div className="flex flex-col">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="border-b border-border-faint py-5 flex flex-row justify-between items-center">
            <Skeleton className="h-4 w-48" />
            <div className="flex flex-row gap-2 items-center">
              <Skeleton className="h-3 w-16" />
              <Skeleton className="h-5 w-14" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
