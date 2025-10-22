import { createContext, useContext, createSignal, createEffect, JSX } from 'solid-js'
import { setGlobalAccessToken, installAuthFetchInterceptor, apiMutate } from '../lib/apiClient'
import { getAuthApiUrl } from '../api/config'

export interface User {
  id: number
  username: string
}

export interface AuthContextType {
  user: () => User | null
  login: (username: string, password: string) => Promise<boolean>
  logout: () => Promise<void>
  isAuthenticated: () => boolean
  isLoading: () => boolean
  getAccessToken: () => string | null
  setAccessToken: (token: string | null) => void
}

const AuthContext = createContext<AuthContextType>()

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

interface AuthProviderProps {
  children: JSX.Element
}

export function AuthProvider(props: AuthProviderProps) {
  const [user, setUser] = createSignal<User | null>(null)
  const [isLoading, setIsLoading] = createSignal(true)
  const [accessToken, setAccessToken] = createSignal<string | null>(null)

  // Ensure global fetch interceptor is installed and update token
  createEffect(() => {
    installAuthFetchInterceptor()
    setGlobalAccessToken(accessToken())
  })

  // Check authentication status on mount using access token or try refresh
  createEffect(async () => {
    setIsLoading(true)
    
    // Try to refresh token on startup (in case we have a valid refresh cookie)
    try {
      const refreshResponse = await fetch(getAuthApiUrl('refreshToken'), {
        method: 'POST',
        credentials: 'include', // Include cookies for refresh token
      })

      if (refreshResponse.ok) {
        const refreshData = await refreshResponse.json()
        if (refreshData.success && refreshData.data.access_token) {
          const newToken = refreshData.data.access_token
          setAccessToken(newToken)
          
          // Get user info with the new access token
          const statusResponse = await fetch(getAuthApiUrl('status'), {
            headers: {
              'Authorization': `Bearer ${newToken}`
            }
          })
          
          if (statusResponse.ok) {
            const statusData = await statusResponse.json()
            if (statusData.success && statusData.data && statusData.data.user) {
              setUser(statusData.data.user)
            }
          }
        }
      }
    } catch (error) {
      console.log('Auto-refresh failed:', error)
    } finally {
      setIsLoading(false)
    }
  })

  const login = async (username: string, password: string): Promise<boolean> => {
    try {
      console.log('Starting login...')
      
      // Use the unified API client for login
      const loginData = await apiMutate<{ access_token: string }>(
        getAuthApiUrl('login'),
        {
          method: 'POST',
          body: { username, password }
        }
      )

      if (loginData.access_token) {
        const newToken = loginData.access_token
        console.log('Got access token, setting it...')
        setAccessToken(newToken)
        
        // Get user info with the access token
        console.log('Making status call...')
        const statusResponse = await fetch(getAuthApiUrl('status'), {
          headers: {
            'Authorization': `Bearer ${newToken}`
          }
        })
        
        console.log('Status response status:', statusResponse.status)
        if (statusResponse.ok) {
          const statusData = await statusResponse.json()
          console.log('Status response data:', statusData)
          if (statusData.success && statusData.data && statusData.data.user) {
            console.log('Setting user:', statusData.data.user)
            setUser(statusData.data.user)
            return true
          }
        }
      }
      console.log('Login failed - returning false')
      return false
    } catch (error) {
      console.error('Login failed:', error)
      return false
    }
  }

  const logout = async (): Promise<void> => {
    try {
      await apiMutate(getAuthApiUrl('logout'), { method: 'POST' })
    } catch (error) {
      console.error('Logout failed:', error)
    } finally {
      setUser(null)
      setAccessToken(null)
    }
  }

  const isAuthenticated = () => user() !== null && accessToken() !== null

  const getAccessToken = () => accessToken()

  const setAccessTokenWrapper = (token: string | null) => {
    setAccessToken(token)
  }

  const authContextValue: AuthContextType = {
    user,
    login,
    logout,
    isAuthenticated,
    isLoading,
    getAccessToken,
    setAccessToken: setAccessTokenWrapper
  }

  return (
    <AuthContext.Provider value={authContextValue}>
      {props.children}
    </AuthContext.Provider>
  )
}