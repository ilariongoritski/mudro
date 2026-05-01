import * as React from 'react'
import { cn } from '@/shared/lib/utils'

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          'flex h-10 w-full rounded-xl border border-mudro-line bg-white/5 px-3 py-2 text-sm text-mudro-text ring-offset-mudro-bg file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-mudro-muted/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-mudro-accent focus-visible:ring-offset-1 focus-visible:border-mudro-accent/50 disabled:cursor-not-allowed disabled:opacity-40 transition-colors',
          className,
        )}
        ref={ref}
        {...props}
      />
    )
  },
)
Input.displayName = 'Input'

export { Input }
