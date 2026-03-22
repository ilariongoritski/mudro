import './Skeleton.css'

interface SkeletonProps {
  type?: 'text' | 'title' | 'circle' | 'rect'
  width?: string | number
  height?: string | number
  className?: string
  style?: React.CSSProperties
}

export const Skeleton = ({
  type = 'text',
  width,
  height,
  className = '',
  style,
}: SkeletonProps) => {
  const classes = ['skeleton', `skeleton-${type}`, className]
    .filter(Boolean)
    .join(' ')

  return (
    <div
      className={classes}
      style={{ width, height, ...style }}
      aria-hidden="true"
    />
  )
}
