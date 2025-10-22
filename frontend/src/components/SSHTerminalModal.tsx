import { createSignal, createEffect, onCleanup, Show } from 'solid-js'
import type { Component } from 'solid-js'
import type { SSHHost } from '../types/remote'

interface SSHTerminalModalProps {
  host: SSHHost | null
  isOpen: boolean
  onClose: () => void
  token?: string
}

const SSHTerminalModal: Component<SSHTerminalModalProps> = (props) => {
  const [ws, setWs] = createSignal<WebSocket | null>(null)
  const [terminalOutput, setTerminalOutput] = createSignal('')

  createEffect(() => {
    if (props.isOpen && props.host) {
      setTerminalOutput('')
      
      // Use relative WebSocket URL to leverage the proxy
      const params = new URLSearchParams({ host_id: String(props.host.uid) })
      if (props.token) params.set('access_token', props.token)
      const websocketUrl = `/api/ssh/connect?${params.toString()}`
      
      const socket = new WebSocket(websocketUrl)
      
      socket.onopen = () => {
        setTerminalOutput(prev => prev + `Connected to ${props.host!.name} (${props.host!.addr})\r\n`)
      }
      
      socket.onmessage = (event) => {
        setTerminalOutput(prev => prev + event.data)
      }
      
      socket.onclose = () => {
        setTerminalOutput(prev => prev + '\r\nConnection closed\r\n')
      }
      
      socket.onerror = (error) => {
        console.error('WebSocket connection error:', error)
        setTerminalOutput(prev => prev + `\r\nConnection failed. Check the server URL and ensure the backend is running on the correct port (e.g., :8285). Details: ${error.type || 'Unknown error'}\r\n`)
      }
      
      setWs(socket)
    } else {
      // Close WebSocket when modal closes
      const socket = ws()
      if (socket) {
        socket.close()
        setWs(null)
      }
    }
  })

  onCleanup(() => {
    const socket = ws()
    if (socket) {
      socket.close()
    }
  })

  return (
    <Show when={props.isOpen}>
      <div class="modal modal-open">
        <div class="modal-box max-w-4xl h-3/4">
          <h3 class="font-bold text-lg mb-4">
            SSH Terminal - {props.host?.name}
          </h3>
          
          <div class="bg-black text-green-400 font-mono text-sm p-4 rounded h-full max-h-96 overflow-y-auto">
            <pre class="whitespace-pre-wrap">{terminalOutput()}</pre>
          </div>

          <div class="modal-action">
            <button 
              class="btn"
              onClick={props.onClose}
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </Show>
  )
}

export default SSHTerminalModal
