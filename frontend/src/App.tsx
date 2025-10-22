import type { Component } from 'solid-js'
import { Router } from '@solidjs/router'
import { AuthProvider } from './contexts/AuthContext'
import { I18nProvider } from './i18n'
import { routes } from './routes'

const App: Component = () => {
  return (
    <I18nProvider>
      <AuthProvider>
        <Router>
          {routes}
        </Router>
      </AuthProvider>
    </I18nProvider>
  )
}

export default App
