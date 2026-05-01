import * as React from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/shared/lib/utils'

const badgeVariants = cva(
  'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold transition-colors',
  {
    variants: {
      variant: {
        // Нейтральный тёмный
        default: 'border border-mudro-line bg-white/8 text-mudro-muted',
        // Платформы
        vk:      'bg-vk/15 text-vk border border-vk/25',
        tg:      'bg-tg/15 text-tg border border-tg/25',
        // Акцент
        accent:  'bg-mudro-accent/15 text-mudro-accent border border-mudro-accent/25',
        // Успех
        success: 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/25',
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
