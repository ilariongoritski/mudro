import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/shared/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold transition-colors',
  {
    variants: {
      variant: {
        default: 'border border-slate-200 bg-white text-slate-700',
        vk: 'bg-vk/10 text-vk border border-vk/20',
        tg: 'bg-tg/10 text-tg border border-tg/20',
        pink: 'bg-mudro-pink/10 text-mudro-pink border border-mudro-pink/20',
        success: 'bg-emerald-50 text-emerald-700 border border-emerald-200',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />
}

export { Badge }
