import { cn } from "@/lib/utils"

function EmptyState({
  title,
  description,
  className,
}: {
  title: string
  description?: string
  className?: string
}) {
  return (
    <div className={cn("flex flex-col justify-center items-center text-center", className)}>
      <h3 className="font-bold p-2">{title}</h3>
      {description && <p className="text-text-faint w-90 p-1 text-sm">{description}</p>}
    </div>
  )
}

export { EmptyState }
