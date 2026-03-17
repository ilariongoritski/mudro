import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[Mudro ErrorBoundary]', error, info.componentStack)
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null })
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback

      return (
        <div className="mudro-error-boundary">
          <div className="mudro-error-boundary__inner">
            <span className="mudro-error-boundary__eyebrow">Ошибка приложения</span>
            <h2>Что-то сломалось</h2>
            <p>Произошла непредвиденная ошибка в интерфейсе. Попробуйте обновить страницу.</p>
            {this.state.error ? (
              <pre className="mudro-error-boundary__detail">{this.state.error.message}</pre>
            ) : null}
            <div className="mudro-error-boundary__actions">
              <button type="button" onClick={this.handleRetry}>
                Попробовать снова
              </button>
              <button type="button" onClick={() => window.location.reload()}>
                Перезагрузить страницу
              </button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
