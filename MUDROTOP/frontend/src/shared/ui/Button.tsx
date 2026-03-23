import type { ButtonHTMLAttributes, PropsWithChildren } from 'react'

type Tone = 'primary' | 'secondary' | 'neutral' | 'link'

type ButtonProps = PropsWithChildren<
  ButtonHTMLAttributes<HTMLButtonElement> & {
    tone?: Tone
  }
>

export const Button = ({ tone = 'neutral', className = '', children, ...props }: ButtonProps) => {
  return (
    <button className={`button button--${tone} ${className}`.trim()} {...props}>
      {children}
    </button>
  )
}
