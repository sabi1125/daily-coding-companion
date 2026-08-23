import type { badgeVariants } from "@/components/ui/badge";
import type { VariantProps } from "class-variance-authority";

function ResolveBadgeVariant(status: string): VariantProps<typeof badgeVariants>["variant"] {
  if (status === "Solved") return "solved"
  if (status === "Failed") return "failed"
  return "open"
}

export default ResolveBadgeVariant
