import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/shared/lib/utils'

const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-xl text-sm font-semibold transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-mudro-accent focus-visible:ring-offset-2 focus-visible:ring-offset-mudro-bg disabled:pointer-events-none disabled:opacity-40 cursor-pointer',
  {
    variants: {
      variant: {
        // Акцентный — розовый с glow
        default:     'bg-mudro-accent text-white hover:opacity-90 shadow-[0_0_20px_var(--mudro-accent-glow)]',
        // Деструктивный
        destructive: 'bg-red-600/80 text-white hover:bg-red-600 border border-red-500/30',
        // Контурный стеклянный
        outline:     'border border-mudro-line bg-transparent text-mudro-text hover:border-mudro-accent hover:text-mudro-accent',
        // Полупрозрачный вторичный
        secondary:   'bg-white/8 text-mudro-text hover:bg-white/12 border border-mudro-line',
        // Призрак без фона
        ghost:       'bg-transparent text-mudro-muted hover:bg-white/6 hover:text-mudro-text',
        // Ссылка
        link:        'text-mudro-accent underline-offset-4 hover:underline p-0 h-auto',
      },
      size: {
        default: 'h-10 px-4 py-2',
        sm:      'h-8 rounded-lg px-3 text-xs',
        lg:      'h-12 rounded-xl px-8 text-base',
        icon:    'h-10 w-10',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button'
    return (
      <Comp className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />
    )
  },
)
Button.displayName = 'Button'

export { Button }
