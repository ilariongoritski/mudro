import './MudroLogoMark.css'

type MudroLogoMarkProps = {
  className?: string
  label?: string
}

export const MudroLogoMark = ({ className, label = 'Mudro' }: MudroLogoMarkProps) => (
  <span className={className ? `mudro-logo-mark ${className}` : 'mudro-logo-mark'} role="img" aria-label={label}>
    <svg
      className="mudro-logo-mark__svg"
      viewBox="0 0 180 124"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
      focusable="false"
    >
      <path
        d="M31 18C16.641 18 5 29.641 5 44v23c0 14.359 11.641 26 26 26h92.824l34.471 24.025-9.744-24.025H149c14.359 0 26-11.641 26-26V44c0-14.359-11.641-26-26-26H31Z"
        fill="#FFFCFE"
        stroke="#110A14"
        strokeWidth="5"
        strokeLinejoin="round"
      />
      <text
        x="89"
        y="66"
        textAnchor="middle"
        dominantBaseline="middle"
        fill="#110A14"
        fontSize="33"
        fontWeight="800"
        letterSpacing="3"
        fontFamily="'Trebuchet MS', 'Arial Rounded MT Bold', Arial, sans-serif"
      >
        МУДРО.
      </text>
    </svg>
  </span>
)
