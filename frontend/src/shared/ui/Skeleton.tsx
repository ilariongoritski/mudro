import { cn } from '@/shared/lib/utils'

interface SkeletonProps {
  type?: 'text' | 'title' | 'circle' | 'rect'
  width?: string | number
  height?: string | number
  className?: string
  style?: React.CSSProperties
}

const typeClasses: Record<NonNullable<SkeletonProps['type']>, string> = {
  text:   'h-4 rounded-lg mb-2',
  title:  'h-6 rounded-lg mb-3',
  circle: 'rounded-full',
  rect:   'rounded-xl',
}

export const Skeleton = ({
  type = 'text',
  width,
  height,
  className,
  style,
}: SkeletonProps) => {
  return (
    <div
      className={cn(
        'w-full animate-pulse bg-white/8',
        typeClasses[type],
        className,
      )}
      style={{ width, height, ...style }}
      aria-hidden="true"
    />
  )
}
