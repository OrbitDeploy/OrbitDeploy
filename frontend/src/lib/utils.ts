// No longer need these imports for browser environment
// import * as os from 'os'
// import * as path from 'path'

/**
 * Strips protocol (http://, https://) from domain string
 * @param input Domain string that may contain protocol
 * @returns Clean domain name without protocol
 */
export function stripProtocolFromDomain(input: string): string {
  if (!input) return ''
  
  // Trim whitespace
  input = input.trim()
  
  // Remove protocol if present
  if (input.startsWith('http://') || input.startsWith('https://')) {
    try {
      const url = new URL(input)
      return url.hostname
    } catch {
      // If URL parsing fails, try simple string replacement
      return input.replace(/^https?:\/\//, '').split('/')[0].split(':')[0]
    }
  }
  
  // Remove port if present
  const colonIndex = input.lastIndexOf(':')
  if (colonIndex !== -1 && !input.includes('[')) {
    // Make sure it's not IPv6
    input = input.substring(0, colonIndex)
  }
  
  return input
}

/**
 * Validates if a string is a valid domain name format
 * @param domain Domain string to validate
 * @returns true if valid domain format
 */
export function isValidDomain(domain: string): boolean {
  if (!domain || domain.length > 253) return false
  
  // Basic domain regex - matches valid domain names
  const domainRegex = /^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$/
  return domainRegex.test(domain)
}

/**
 * Converts a Docker command to Quadlet file format
 * @param dockerCommand The docker run command to convert
 * @returns Object containing quadlet content and env file content
 */
export function dockerCommandToQuadlet(dockerCommand: string): { quadletContent: string; envFileContent: string } {
  // Remove 'docker run' prefix and normalize command
  // Handle backslashes by removing them and joining continuation lines
  const normalizedCommand = dockerCommand
    .replace(/^docker\s+run\s*/i, '')
    .replace(/\s*\\\s*\n\s*/g, ' ') // Remove backslash line continuations
    .replace(/\s*\\\s*/g, ' ') // Remove remaining backslashes
    .trim()

  // Initialize quadlet sections
  const unitSection = ['[Unit]', 'Description=Container converted from Docker command', 'After=network-online.target', '']
  const containerSection = ['[Container]']
  const installSection = ['', '[Install]', 'WantedBy=default.target']

  // Parse command line arguments
  const args = parseDockerCommand(normalizedCommand)

  // Extract image name (last non-flag argument or explicitly specified)
  let imageName = ''
  if (args.image) {
    imageName = args.image
  } else {
    // Find the last argument that doesn't start with '-' and isn't a value for a flag
    const nonFlagArgs = args.remaining.filter(arg => !arg.startsWith('-'))
    if (nonFlagArgs.length > 0) {
      imageName = nonFlagArgs[nonFlagArgs.length - 1]
    }
  }

  if (imageName) {
    containerSection.push(`Image=${imageName}`)
  }

  // Add auto-update if specified
  if (args.autoUpdate) {
    containerSection.push('AutoUpdate=registry')
  }

  // Add execution command if specified
  if (args.command) {
    containerSection.push(`Exec=${args.command}`)
  }

  // Add port mappings
  if (args.ports && args.ports.length > 0) {
    args.ports.forEach(port => {
      containerSection.push(`PublishPort=${port}`)
    })
  }

  // Add volume mappings
  if (args.volumes && args.volumes.length > 0) {
    args.volumes.forEach(volume => {
      containerSection.push(`Volume=${volume}`)
    })
  }

  // Handle environment variables: if present, create env file instead of inline Environment entries
  let envFileContent = ''
  if (args.envVars && args.envVars.length > 0) {
    envFileContent = args.envVars.join('\n')
    // Always use default env file path when env vars are present
    const envFilePath = args.envFile || getDefaultEnvFilePath()
    if (envFilePath) {
      containerSection.push(`EnvironmentFile=${envFilePath}`)
    }
  } else if (args.envFile) {
    // Only add environment file if specified and no env vars to convert
    containerSection.push(`EnvironmentFile=${args.envFile}`)
  }

  // Combine all sections
  const quadletContent = [
    ...unitSection,
    ...containerSection,
    ...installSection
  ].join('\n')

  return { quadletContent, envFileContent }
}

interface ParsedDockerArgs {
  image?: string
  ports: string[]
  volumes: string[]
  envFile?: string
  envVars: string[]
  command?: string
  autoUpdate?: boolean
  remaining: string[]
}

/**
 * Parses docker command arguments
 */
function parseDockerCommand(command: string): ParsedDockerArgs {
  const result: ParsedDockerArgs = {
    ports: [],
    volumes: [],
    envVars: [],
    remaining: []
  }

  // Split command into tokens, handling quoted strings
  const tokens = tokenizeCommand(command)
  
  for (let i = 0; i < tokens.length; i++) {
    const token = tokens[i]
    
    // Port mapping: -p or --publish
    if (token === '-p' || token === '--publish') {
      const portMapping = tokens[++i]
      if (portMapping) {
        result.ports.push(portMapping)
      }
    }
    // Volume mapping: -v or --volume
    else if (token === '-v' || token === '--volume') {
      const volumeMapping = tokens[++i]
      if (volumeMapping) {
        result.volumes.push(volumeMapping)
      }
    }
    // Environment file: --env-file
    else if (token === '--env-file') {
      result.envFile = tokens[++i]
    }
    // Environment variable: -e or --env
    else if (token === '-e' || token === '--env') {
      const envVar = tokens[++i]
      if (envVar) {
        result.envVars.push(envVar)
      }
    }
    // Auto-update detection (custom logic)
    else if (token === '--pull' && tokens[i + 1] === 'always') {
      result.autoUpdate = true
      i++ // skip 'always'
    }
    // Image name pattern detection
    else if (!token.startsWith('-') && isImageName(token)) {
      result.image = token
    }
    // Command after image
    else if (result.image && !token.startsWith('-')) {
      // Everything after image name is the command
      result.command = tokens.slice(i).join(' ')
      break
    }
    else {
      result.remaining.push(token)
    }
  }

  return result
}

/**
 * Tokenizes a command string, respecting quoted strings
 */
function tokenizeCommand(command: string): string[] {
  const tokens: string[] = []
  let current = ''
  let inQuotes = false
  let quoteChar = ''

  for (let i = 0; i < command.length; i++) {
    const char = command[i]
    
    if ((char === '"' || char === "'") && !inQuotes) {
      inQuotes = true
      quoteChar = char
    } else if (char === quoteChar && inQuotes) {
      inQuotes = false
      quoteChar = ''
    } else if (char === ' ' && !inQuotes) {
      if (current.trim()) {
        tokens.push(current.trim())
        current = ''
      }
    } else {
      current += char
    }
  }
  
  if (current.trim()) {
    tokens.push(current.trim())
  }
  
  return tokens
}

/**
 * Checks if a token looks like a Docker image name
 */
function isImageName(token: string): boolean {
  // Basic pattern for Docker image names
  // Examples: nginx, nginx:latest, ghcr.io/owner/image:tag
  const imagePattern = /^[a-zA-Z0-9][a-zA-Z0-9._-]*(?:\/[a-zA-Z0-9][a-zA-Z0-9._-]*)*(?::[a-zA-Z0-9][a-zA-Z0-9._-]*)?$/
  return imagePattern.test(token)
}

/**
 * Default env-file location under the current user's home directory
 */
function getDefaultEnvFilePath(): string {
  // Return a standard path that works across environments
  return '$HOME/.config/.env'
}

/**
 * Extracts the host port from PublishPort directive in Quadlet content
 * PublishPort format: "host_port:container_port" (e.g., "9092:9090")
 * Returns the host port (the first number before the colon)
 */
export function extractHostPortFromQuadlet(quadletContent: string): number | null {
  if (!quadletContent) return null

  const lines = quadletContent.split('\n')
  for (const line of lines) {
    const trimmed = line.trim()
    if (trimmed.startsWith('PublishPort=')) {
      const publishPort = trimmed.substring('PublishPort='.length).trim()
      const parts = publishPort.split(':')
      if (parts.length >= 2) {
        const hostPort = parseInt(parts[0].trim(), 10)
        if (!isNaN(hostPort)) {
          return hostPort
        }
      }
    }
  }
  
  return null
}