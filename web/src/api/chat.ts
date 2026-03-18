import { ApiError, apiBaseUrl, apiRequest } from '@/lib/api'

export interface UserModel {
  id: string
  model_key: string
  display_name: string
  provider_type: string
}

export interface ConversationSummary {
  id: string
  title: string
  model_id: string | null
  message_count: number
  status: string
}

export interface MessageItem {
  id: string
  role: string
  content: string
  reasoning_content: string
  status: string
  finish_reason?: string | null
  usage?: Record<string, unknown>
  created_at: string
}

interface StreamStartEvent {
  type: 'response.start'
  request_id: string
  conversation_id: string
  assistant_message_id: string
}

interface StreamDeltaEvent {
  type: 'response.delta'
  delta: {
    content?: string
  }
}

interface StreamReasoningEvent {
  type: 'reasoning.delta'
  delta: {
    reasoning_content?: string
  }
}

interface StreamCompletedEvent {
  type: 'response.completed'
  request_id?: string
  conversation_id?: string
  assistant_message_id?: string
  finish_reason?: string
  usage?: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
}

interface StreamFailedEvent {
  type: 'response.failed'
  request_id?: string
  conversation_id?: string
  assistant_message_id?: string
  error?: {
    code?: string
    message?: string
  }
}

export type StreamEvent =
  | StreamStartEvent
  | StreamDeltaEvent
  | StreamReasoningEvent
  | StreamCompletedEvent
  | StreamFailedEvent

export async function listModels(accessToken: string) {
  const { data } = await apiRequest<UserModel[]>('/models', {
    accessToken
  })
  return data
}

export async function listConversations(accessToken: string) {
  const { data } = await apiRequest<ConversationSummary[]>('/conversations', {
    accessToken
  })
  return data
}

export async function createConversation(accessToken: string, body: { title: string; model_id: string | null }) {
  const { data } = await apiRequest<ConversationSummary>('/conversations', {
    method: 'POST',
    accessToken,
    body
  })
  return data
}

export async function listConversationMessages(accessToken: string, conversationID: string) {
  const { data } = await apiRequest<MessageItem[]>(`/conversations/${conversationID}/messages`, {
    accessToken
  })
  return data
}

export async function streamChatCompletion(
  body: Record<string, unknown>,
  accessToken: string,
  signal: AbortSignal,
  onEvent: (event: StreamEvent) => void
) {
  const response = await fetch(`${apiBaseUrl}/api/v1/chat/completions`, {
    method: 'POST',
    headers: {
      Accept: 'text/event-stream',
      'Content-Type': 'application/json',
      Authorization: `Bearer ${accessToken}`
    },
    credentials: 'include',
    body: JSON.stringify(body),
    signal
  })

  if (!response.ok || !response.body) {
    throw await toStreamError(response)
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    const chunks = buffer.split('\n\n')
    buffer = chunks.pop() ?? ''

    for (const chunk of chunks) {
      const event = parseSSEChunk(chunk)
      if (event === '[DONE]') {
        return
      }
      if (event) {
        onEvent(event)
      }
    }
  }

  buffer += decoder.decode()
  const lastEvent = parseSSEChunk(buffer)
  if (lastEvent && lastEvent !== '[DONE]') {
    onEvent(lastEvent)
  }
}

function parseSSEChunk(chunk: string): StreamEvent | '[DONE]' | null {
  const lines = chunk
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.startsWith('data:'))

  if (lines.length === 0) {
    return null
  }

  const payload = lines.map((line) => line.slice(5).trim()).join('\n')
  if (!payload) {
    return null
  }
  if (payload === '[DONE]') {
    return '[DONE]'
  }

  return JSON.parse(payload) as StreamEvent
}

async function toStreamError(response: Response) {
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    const payload = (await response.json()) as {
      error?: {
        code?: string
        message?: string
        details?: unknown
      }
    }
    return new ApiError(
      payload.error?.message ?? `Request failed with status ${response.status}`,
      payload.error?.code ?? 'HTTP_ERROR',
      response.status,
      payload.error?.details
    )
  }

  return new ApiError(`Request failed with status ${response.status}`, 'HTTP_ERROR', response.status)
}
