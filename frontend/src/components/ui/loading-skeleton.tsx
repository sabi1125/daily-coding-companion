import { Skeleton } from "@/components/ui/skeleton"

export function LoadingSkeleton() {
  return (
    <div className="mx-auto w-full max-w-5xl px-10 py-14 flex flex-col gap-5">
      {/* Badges */}

      <div className="flex flex-row justify-between items-center">
        <div className="flex flex-row gap-1">
          <Skeleton className="h-6 w-20" />
          <Skeleton className="h-6 w-24" />
          <Skeleton className="h-6 w-16" />
        </div>
        <div className="flex flex-row gap-2">
          <Skeleton className="h-6 w-16" />
          <Skeleton className="h-6 w-20" />
        </div>
      </div>

      {/* Title + problem text */}

      <Skeleton className="h-8 w-80" />
      <div className="flex flex-col gap-2">
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-2/3" />
      </div>

      <Skeleton className="h-4 w-32" />

      {/* Editor */}

      <div className="rounded-lg border border-border-faint overflow-hidden shadow-2xs">
        <div className="flex flex-row justify-between items-center h-9 p-4 border-b border-border-faint">
          <Skeleton className="h-4 w-16" />
          <Skeleton className="h-6 w-48" />
        </div>
        <Skeleton className="h-[450px] w-full rounded-none" />
      </div>

      {/* Toggle group + submit */}

      <div className="flex flex-row justify-between items-center">
        <Skeleton className="h-7 w-32 rounded-lg" />
        <Skeleton className="h-8 w-28" />
      </div>
    </div>
  )
}
