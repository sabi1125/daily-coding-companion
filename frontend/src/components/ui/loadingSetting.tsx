import { Skeleton } from "@/components/ui/skeleton"

export function LoadingSetting() {
  return (
    <div className="mx-auto w-full max-w-2xl px-10 py-14 flex flex-col gap-8">
      {/* Header */}

      <Skeleton className="h-8 w-32" />

      {/* Account */}

      <div className="gap-2 flex flex-col">
        <Skeleton className="h-4 w-20" />
        <div className="rounded-lg border border-border-faint p-4 flex justify-between items-center gap-2">
          <Skeleton className="h-4 w-40" />
          <Skeleton className="h-8 w-20" />
        </div>
        <Skeleton className="h-3 w-64" />
      </div>

      {/* Daily Ingest */}

      <div className="gap-2 flex flex-col">
        <Skeleton className="h-4 w-28" />
        <div className="rounded-lg border border-border-faint p-4 flex justify-between items-center gap-4">
          <div className="flex flex-col gap-2">
            <Skeleton className="h-4 w-48" />
            <Skeleton className="h-3 w-72" />
          </div>
          <Skeleton className="h-6 w-20 shrink-0" />
        </div>
      </div>

      {/* Stats */}

      <div className="flex flex-col gap-2">
        <Skeleton className="h-4 w-16" />
        <div className="grid grid-cols-2 gap-3">
          <div className="rounded-lg border border-border-faint p-4 flex flex-col gap-1">
            <Skeleton className="h-8 w-10" />
            <Skeleton className="h-4 w-16" />
          </div>
          <div className="rounded-lg border border-border-faint p-4 flex flex-col gap-1">
            <Skeleton className="h-8 w-10" />
            <Skeleton className="h-4 w-16" />
          </div>
        </div>
      </div>

      {/* Preferences */}

      <div className="gap-2 flex flex-col">
        <Skeleton className="h-4 w-36" />
        <Skeleton className="h-3 w-full" />
        <Skeleton className="h-20 w-full rounded-lg" />
        <Skeleton className="h-8 w-20" />
      </div>
    </div>
  )
}
