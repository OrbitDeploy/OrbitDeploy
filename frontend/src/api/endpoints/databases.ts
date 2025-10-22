// Database API endpoints
import type { ApiEndpoint } from '../apiHooksW'

export const listDatabasesEndpoint = (): ApiEndpoint => ({
  url: '/api/databases',
  method: 'GET',
})

export const createDatabaseEndpoint = (): ApiEndpoint => ({
  url: '/api/databases',
  method: 'POST',
})

export const getDatabaseEndpoint = (uid: string): ApiEndpoint => ({
  url: `/api/databases/${uid}`,
  method: 'GET',
})

export const updateDatabaseEndpoint = (uid: string): ApiEndpoint => ({
  url: `/api/databases/${uid}`,
  method: 'PATCH',
})

export const deleteDatabaseEndpoint = (uid: string): ApiEndpoint => ({
  url: `/api/databases/${uid}`,
  method: 'DELETE',
})

export const deployDatabaseEndpoint = (uid: string): ApiEndpoint => ({
  url: `/api/databases/${uid}/deploy`,
  method: 'POST',
})

export const startDatabaseEndpoint = (uid: string): ApiEndpoint => ({
  url: `/api/databases/${uid}/start`,
  method: 'POST',
})

export const stopDatabaseEndpoint = (uid: string): ApiEndpoint => ({
  url: `/api/databases/${uid}/stop`,
  method: 'POST',
})

export const restartDatabaseEndpoint = (uid: string): ApiEndpoint => ({
  url: `/api/databases/${uid}/restart`,
  method: 'POST',
})

export const getDatabaseConnectionInfoEndpoint = (uid: string, unmask = false): ApiEndpoint => ({
  url: `/api/databases/${uid}/connection-info${unmask ? '?unmask=true' : ''}`,
  method: 'GET',
})
