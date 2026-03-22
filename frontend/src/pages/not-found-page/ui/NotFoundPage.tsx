import { Link } from 'react-router-dom'
// We can just use the ErrorBoundary styles for consistency!
// We can just use the ErrorBoundary styles for consistency!

export const NotFoundPage = () => {
  return (
    <div className="mudro-error-boundary">
      <div className="mudro-error-boundary__inner">
        <span className="mudro-error-boundary__eyebrow">404</span>
        <h2>Страница не найдена</h2>
        <p>Возможно, вы ввели неправильный адрес, или страница была удалена.</p>
        
        <div className="mudro-error-boundary__actions">
          <Link to="/">
            <button type="button">На главную</button>
          </Link>
        </div>
      </div>
    </div>
  )
}
