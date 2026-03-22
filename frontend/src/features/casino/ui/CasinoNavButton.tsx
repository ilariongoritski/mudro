import { Link } from 'react-router-dom'

import './CasinoNavButton.css'

export const CasinoNavButton = () => {
  return (
    <Link className="casino-nav-button" to="/casino" aria-label="Открыть раздел казино">
      <span className="casino-nav-button__chip">NEW</span>
      <span className="casino-nav-button__label">Казино</span>
    </Link>
  )
}
