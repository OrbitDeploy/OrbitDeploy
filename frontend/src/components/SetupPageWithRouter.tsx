import { Component } from 'solid-js'
import { useNavigate } from '@solidjs/router'
import SetupPageComponent from '../pages/SetupPage'

const SetupPageWithRouter: Component = () => {
  const navigate = useNavigate()

  const handleSetupComplete = () => {
    // Reload the page to refresh auth state and redirect to dashboard
    window.location.href = '/'
  }

  return <SetupPageComponent onSetupComplete={handleSetupComplete} />
}

export default SetupPageWithRouter