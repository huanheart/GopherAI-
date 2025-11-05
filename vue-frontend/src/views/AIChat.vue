<template>
  <div class="ai-chat-container">
    <!-- 左侧会话列表 -->
    <div class="session-list">
      <div class="session-list-header">
        <span>会话列表</span>
        <button class="new-chat-btn" @click="createNewSession">＋ 新聊天</button>
      </div>
      <ul class="session-list-ul">
        <li
          v-for="session in sessions"
          :key="session.id"
          :class="['session-item', { active: currentSessionId === session.id }]"
          @click="switchSession(session.id)"
        >
          {{ session.name || `会话 ${session.id}` }}
        </li>
      </ul>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-section">
      <div class="top-bar">
        <button class="back-btn" @click="$router.push('/menu')">← 返回</button>
        <button class="sync-btn" @click="syncHistory" :disabled="!currentSessionId || tempSession">同步历史数据</button>
        <label for="modelType">选择模型：</label>
        <select id="modelType" v-model="selectedModel" class="model-select">
          <option value="1">阿里百炼</option>
          <option value="2">豆包</option>
          <option value="3">百炼RAG</option>
          <option value="4">阿里百炼MCP</option>
        </select>
        <label for="streamingMode" style="margin-left: 20px;">
          <input type="checkbox" id="streamingMode" v-model="isStreaming" />
          流式响应
        </label>
      </div>

      <div class="chat-messages" ref="messagesRef">
        <div
          v-for="(message, index) in currentMessages"
          :key="index"
          :class="['message', message.role === 'user' ? 'user-message' : 'ai-message']"
        >
          <div class="message-header">
            <b>{{ message.role === 'user' ? '你' : 'AI' }}:</b>
            <button v-if="message.role === 'assistant'" class="tts-btn" @click="playTTS(message.content)">🔊</button>
            <span v-if="message.meta && message.meta.status === 'streaming'" class="streaming-indicator"> ··</span>
          </div>
          <div class="message-content" v-html="renderMarkdown(message.content)"></div>
        </div>
      </div>

      <div class="chat-input">
        <textarea
          v-model="inputMessage"
          placeholder="请输入你的问题..."
          @keydown.enter.exact.prevent="sendMessage"
          :disabled="loading"
          ref="messageInput"
          rows="1"
        ></textarea>
        <button
          type="button"
          :disabled="!inputMessage.trim() || loading"
          @click="sendMessage"
          class="send-btn"
        >
          {{ loading ? '发送中...' : '发送' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
/*
  完整 AIChat.vue
  - 保持你指定的接口路径不变
  - 修复两个问题：
    1) 流式调用：首次新会话走 /api/AI/chat/send-stream-new-session，已有会话走 /api/AI/chat/send-stream
    2) 逐字（逐小片段）优雅渲染：smoothAppend 实现，避免一次性完整渲染
  说明：后端返回格式未强制要求，这里兼容常见三种流数据形式：
    - 纯文本 data: some text
    - JSON 包含 {"delta":"..."} 或 {"sessionId":"..."}（我们会尝试解析）
    - [DONE] 标志结束
*/

import { ref, nextTick, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'AIChat',
  setup() {
    // === reactive state ===
    const sessions = ref({})               // map: sessionId -> {id, name, messages: []}
    const currentSessionId = ref(null)     // string or 'temp'
    const tempSession = ref(false)         // true when user created a temp new session (not persisted)
    const currentMessages = ref([])        // array used for rendering (detached copy)
    const inputMessage = ref('')
    const loading = ref(false)
    const messagesRef = ref(null)
    const messageInput = ref(null)
    const selectedModel = ref('1')
    const isStreaming = ref(false)

    // === smooth render helpers ===
    // We'll keep per-ai-message render buffers so multiple concurrent streams (theoretically) won't collide.
    // Map from aiMessageId (we'll use index or generated id) -> { buffer: string, timer: number|null }
    const _renderBuffers = new Map()

    function _startSmoothTimer(id, aiMessage, speedMs = 30, chunkSize = 2) {
      // if timer exists, do nothing
      const entry = _renderBuffers.get(id)
      if (!entry) return
      if (entry.timer) return

      entry.timer = setInterval(async () => {
        if (!entry.buffer || entry.buffer.length === 0) {
          // no buffer, stop timer
          clearInterval(entry.timer)
          entry.timer = null
          // if stream finished and no buffer, ensure final flush (nothing to flush)
          return
        }
        // Take small chunk and append
        const chunk = entry.buffer.slice(0, chunkSize)
        entry.buffer = entry.buffer.slice(chunk.length)
        // mutate aiMessage content safely
        aiMessage.content += chunk
        // After adding chunk, scroll to bottom
        await nextTick()
        scrollToBottom()
      }, speedMs)
    }

    function _appendToBuffer(id, text, aiMessage, speedMs = 30, chunkSize = 2) {
      let entry = _renderBuffers.get(id)
      if (!entry) {
        entry = { buffer: '', timer: null }
        _renderBuffers.set(id, entry)
      }
      entry.buffer += text
      // start timer if not running
      _startSmoothTimer(id, aiMessage, speedMs, chunkSize)
    }

    function _clearBuffer(id) {
      const entry = _renderBuffers.get(id)
      if (!entry) return
      if (entry.timer) {
        clearInterval(entry.timer)
        entry.timer = null
      }
      _renderBuffers.delete(id)
    }

    // === utility functions ===
    const renderMarkdown = (text) => {
      if (!text && text !== '') return ''
      return String(text)
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/`(.*?)`/g, '<code>$1</code>')
        .replace(/\n/g, '<br>')
    }

    const playTTS = async (text) => {
      try {
        const response = await api.post('/chat/tts', { text })
        if (response.data && response.data.status_code === 1000 && response.data.url) {
          const audio = new Audio(response.data.url)
          audio.play()
        } else {
          ElMessage.error('无法获取语音')
        }
      } catch (error) {
        console.error('TTS error:', error)
        ElMessage.error('请求语音接口失败')
      }
    }

    const loadSessions = async () => {
      try {
        const response = await api.get('/AI/chat/sessions')
        if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.sessions)) {
          const sessionMap = {}
          response.data.sessions.forEach(s => {
            const sid = String(s.sessionId)
            sessionMap[sid] = {
              id: sid,
              name: s.name || `会话 ${sid}`,
              messages: [] // lazy load
            }
          })
          sessions.value = sessionMap
        }
      } catch (error) {
        console.error('Load sessions error:', error)
      }
    }

    const createNewSession = () => {
      currentSessionId.value = 'temp'
      tempSession.value = true
      currentMessages.value = []
      // focus input
      nextTick(() => {
        if (messageInput.value) messageInput.value.focus()
      })
    }

    const switchSession = async (sessionId) => {
      if (!sessionId) return
      currentSessionId.value = String(sessionId)
      tempSession.value = false

      // lazy load history if not present
      if (!sessions.value[sessionId].messages || sessions.value[sessionId].messages.length === 0) {
        try {
          const response = await api.post('/AI/chat/history', { sessionId: currentSessionId.value })
          if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.history)) {
            const messages = response.data.history.map(item => ({
              role: item.is_user ? 'user' : 'assistant',
              content: item.content
            }))
            sessions.value[sessionId].messages = messages
          }
        } catch (err) {
          console.error('Load history error:', err)
        }
      }

      // detach copy
      currentMessages.value = [...(sessions.value[sessionId].messages || [])]
      await nextTick()
      scrollToBottom()
    }

    const syncHistory = async () => {
      if (!currentSessionId.value || tempSession.value) {
        ElMessage.warning('请选择已有会话进行同步')
        return
      }
      try {
        const response = await api.post('/AI/chat/history', { sessionId: currentSessionId.value })
        if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.history)) {
          const messages = response.data.history.map(item => ({
            role: item.is_user ? 'user' : 'assistant',
            content: item.content
          }))
          sessions.value[currentSessionId.value].messages = messages
          currentMessages.value = [...messages]
          await nextTick()
          scrollToBottom()
        } else {
          ElMessage.error('无法获取历史数据')
        }
      } catch (err) {
        console.error('Sync history error:', err)
        ElMessage.error('请求历史数据失败')
      }
    }

    // === send message entry ===
    const sendMessage = async () => {
      if (!inputMessage.value || !inputMessage.value.trim()) {
        ElMessage.warning('请输入消息内容')
        return
      }

      const userMessage = {
        role: 'user',
        content: inputMessage.value
      }
      const currentInput = inputMessage.value
      inputMessage.value = ''

      // push user message into currentMessages (view)
      currentMessages.value.push(userMessage)
      await nextTick()
      scrollToBottom()

      try {
        loading.value = true
        if (isStreaming.value) {
          // streaming flow
          await handleStreaming(currentInput)
        } else {
          // normal non-streaming
          await handleNormal(currentInput)
        }
      } catch (err) {
        console.error('Send message error:', err)
        ElMessage.error('发送失败，请重试')
        // rollback view
        if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value] && sessions.value[currentSessionId.value].messages) {
          // If we were supposed to push to sessions, pop
          const sessionArr = sessions.value[currentSessionId.value].messages
          if (sessionArr && sessionArr.length) sessionArr.pop()
        }
        currentMessages.value.pop()
      } finally {
        if (!isStreaming.value) {
          loading.value = false
        }
        await nextTick()
        scrollToBottom()
      }
    }

    // === streaming handler ===
    async function handleStreaming(question) {
      // Create ai message placeholder and meta
      const aiMessage = {
        role: 'assistant',
        content: '',
        meta: { status: 'streaming' } // mark streaming
      }
      // We'll use an id for buffer map; use timestamp + random
      const aiId = `ai-${Date.now()}-${Math.floor(Math.random() * 10000)}`

      // push placeholder into currentMessages (view)
      currentMessages.value.push(aiMessage)
      // If continue existing session, also push into sessions' message array immediately to keep sync
      if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value]) {
        if (!sessions.value[currentSessionId.value].messages) sessions.value[currentSessionId.value].messages = []
        sessions.value[currentSessionId.value].messages.push({ role: 'assistant', content: '' })
      }

      // Choose URL based on tempSession and keep EXACT paths you required
      const url = tempSession.value
        ? '/api/AI/chat/send-stream-new-session'  // first time new session streaming
        : '/api/AI/chat/send-stream'              // subsequent streaming in existing session

      const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      }

      const body = tempSession.value
        ? { question: question, modelType: selectedModel.value }
        : { question: question, modelType: selectedModel.value, sessionId: currentSessionId.value }

      // fetch + stream reading
      const response = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify(body)
      })

      if (!response.ok) {
        // network/HTTP error
        loading.value = false
        throw new Error('Network response was not ok')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = '' // keep incomplete line
      // we'll update the aiMessage.content via smooth append to avoid flash

      // initialize buffer entry for this aiId
      _renderBuffers.set(aiId, { buffer: '', timer: null })

      async function processChunk(chunkText) {
        // chunkText may include several SSE lines or partial
        buffer += chunkText
        const lines = buffer.split('\n')
        buffer = lines.pop() // leftover

        for (const rawLine of lines) {
          const line = rawLine.trim()
          if (!line) continue

          if (line.startsWith('event:')) {
            const event = line.slice(6).trim()
            // handle event lines
            if (event === 'session') {
              // next line should be data with sessionId
              continue
            } else if (event === 'end') {
              // stream ended
              aiMessage.meta = { status: 'done' }
              return { doneSignal: true }
            }
          } else if (line.startsWith('data:')) {
            const data = line.slice(5).trim()
            // try to parse as JSON first
            try {
              const parsed = JSON.parse(data)
              if (parsed.sessionId) {
                // this is session event data
                const newSid = String(parsed.sessionId)
                if (tempSession.value) {
                  sessions.value[newSid] = {
                    id: newSid,
                    name: sessions.value[newSid]?.name || '新会话',
                    messages: [...currentMessages.value]
                  }
                  currentSessionId.value = newSid
                  tempSession.value = false
                }
              } else if (parsed.message && parsed.message === 'Stream ended') {
                // end event
                aiMessage.meta = { status: 'done' }
                return { doneSignal: true }
              }
            } catch (e) {
              // not JSON, treat as plain text response
              _appendToBuffer(aiId, data, aiMessage)
            }
          }
        }

        return { doneSignal: false }
      }

      // read loop
      try {
        // eslint-disable-next-line no-constant-condition
        while (true) {
          const { done, value } = await reader.read()
          if (done) {
            // stream ended; mark done
            loading.value = false
            // allow remaining buffer to flush via timer
            // finalize: remove streaming meta after short delay to allow last chars to render
            setTimeout(() => {
              aiMessage.meta = { status: 'done' }
              // synchronize with sessions store: if we have an active session, update stored messages
              if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value]) {
                // find last assistant message in sessions and set its content to current content
                const sessMsgs = sessions.value[currentSessionId.value].messages
                if (Array.isArray(sessMsgs) && sessMsgs.length) {
                  // last assistant in sessMsgs likely corresponds to this ai placeholder; update
                  let lastIndex = sessMsgs.length - 1
                  // ensure it's assistant role
                  if (sessMsgs[lastIndex] && sessMsgs[lastIndex].role === 'assistant') {
                    sessMsgs[lastIndex].content = aiMessage.content
                  } else {
                    // push if not present
                    sessMsgs.push({ role: 'assistant', content: aiMessage.content })
                  }
                }
              } else if (!tempSession.value && currentSessionId.value) {
                // If session exists but messages array empty, set it
                if (!sessions.value[currentSessionId.value].messages) {
                  sessions.value[currentSessionId.value].messages = [...currentMessages.value]
                }
              }
              // cleanup buffer entry
              _clearBuffer(aiId)
            }, 350)
            break
          } else {
            const chunkText = decoder.decode(value, { stream: true })
            await processChunk(chunkText)
          }
        }
      } catch (err) {
        console.error('Stream reading error:', err)
        loading.value = false
        aiMessage.meta = { status: 'error' }
        _clearBuffer(aiId)
      } finally {
        // ensure we set loading false for streaming case
        loading.value = false
      }
    }

    // === non-streaming handler ===
    async function handleNormal(question) {
      if (tempSession.value) {
        // create new session with single request
        const response = await api.post('/AI/chat/send-new-session', {
          question: question,
          modelType: selectedModel.value
        })
        if (response.data && response.data.status_code === 1000) {
          const sessionId = String(response.data.sessionId)
          const aiMessage = {
            role: 'assistant',
            content: response.data.Information || ''
          }
          // create session entry
          sessions.value[sessionId] = {
            id: sessionId,
            name: '新会话',
            messages: [ { role: 'user', content: question }, aiMessage ]
          }
          currentSessionId.value = sessionId
          tempSession.value = false
          currentMessages.value = [...sessions.value[sessionId].messages]
        } else {
          ElMessage.error(response.data?.status_msg || '发送失败')
          // rollback
          currentMessages.value.pop()
        }
      } else {
        // continue existing session
        const sessionMsgs = sessions.value[currentSessionId.value].messages
        // push user into persistent session array
        sessionMsgs.push({ role: 'user', content: question })
        // call send
        const response = await api.post('/AI/chat/send', {
          question: question,
          modelType: selectedModel.value,
          sessionId: currentSessionId.value
        })
        if (response.data && response.data.status_code === 1000) {
          const aiMessage = { role: 'assistant', content: response.data.Information || '' }
          sessionMsgs.push(aiMessage)
          currentMessages.value = [...sessionMsgs]
        } else {
          ElMessage.error(response.data?.status_msg || '发送失败')
          sessionMsgs.pop() // rollback
          currentMessages.value.pop()
        }
      }
    }

    // === scrolling ===
    const scrollToBottom = () => {
      if (messagesRef.value) {
        try {
          messagesRef.value.scrollTop = messagesRef.value.scrollHeight
        } catch (e) {
          // ignore
        }
      }
    }

    onMounted(() => {
      loadSessions()
    })

    // expose to template
    return {
      sessions: computed(() => Object.values(sessions.value)),
      currentSessionId,
      tempSession,
      currentMessages,
      inputMessage,
      loading,
      messagesRef,
      messageInput,
      selectedModel,
      isStreaming,
      renderMarkdown,
      playTTS,
      createNewSession,
      switchSession,
      syncHistory,
      sendMessage
    }
  }
}
</script>

<style scoped>
.ai-chat-container {
  min-height: 100vh;
  display: flex;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial;
  color: #222;
}

.ai-chat-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="2" fill="rgba(255,255,255,0.08)"/><circle cx="80" cy="80" r="2" fill="rgba(255,255,255,0.08)"/><circle cx="40" cy="60" r="1" fill="rgba(255,255,255,0.06)"/><circle cx="60" cy="30" r="1.5" fill="rgba(255,255,255,0.06)"/></svg>');
  animation: float 20s ease-in-out infinite;
  opacity: 0.25;
}

@keyframes float {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(180deg); }
}

.session-list {
  width: 280px;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(15px);
  border-right: 1px solid rgba(0, 0, 0, 0.08);
  box-shadow: 2px 0 20px rgba(0, 0, 0, 0.08);
  position: relative;
  z-index: 2;
}

.session-list-header {
  padding: 20px;
  text-align: center;
  font-weight: 600;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.06) 0%, rgba(103, 194, 58, 0.06) 100%);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  display: flex;
  flex-direction: column;
  gap: 12px;
  align-items: center;
}

.new-chat-btn {
  width: 100%;
  padding: 12px 0;
  cursor: pointer;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.28);
  transition: all 0.25s ease;
  position: relative;
  overflow: hidden;
}

.new-chat-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.12), transparent);
  transition: left 0.5s;
}

.new-chat-btn:hover::before {
  left: 100%;
}

.new-chat-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.36);
}

.session-list-ul {
  list-style: none;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow-y: auto;
}

.session-item {
  padding: 15px 20px;
  cursor: pointer;
  border-bottom: 1px solid rgba(0, 0, 0, 0.03);
  transition: all 0.2s ease;
  position: relative;
  color: #2c3e50;
}

.session-item.active {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
  box-shadow: inset 0 0 20px rgba(102, 126, 234, 0.2);
}

.session-item:hover {
  background: rgba(102, 126, 234, 0.06);
  transform: translateX(4px);
}

/* chat section */
.chat-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  z-index: 1;
  min-width: 0;
}

.top-bar {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  color: #2c3e50;
  display: flex;
  align-items: center;
  padding: 12px 24px;
  box-shadow: 0 2px 14px rgba(0, 0, 0, 0.06);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  gap: 12px;
}

.back-btn {
  background: rgba(255, 255, 255, 0.22);
  border: 1px solid rgba(0, 0, 0, 0.06);
  color: #2c3e50;
  padding: 8px 14px;
  border-radius: 10px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.32);
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.08);
}

.sync-btn {
  background: linear-gradient(135deg, #67c23a 0%, #409eff 100%);
  color: white;
  padding: 8px 14px;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(103, 194, 58, 0.2);
  transition: all 0.2s ease;
}

.sync-btn:disabled {
  background: #ccc;
  box-shadow: none;
  cursor: not-allowed;
}

.model-select {
  margin-left: 6px;
  padding: 6px 10px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  border-radius: 8px;
  background: white;
  color: #2c3e50;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 30px;
  display: flex;
  flex-direction: column;
  gap: 18px;
  position: relative;
  z-index: 1;
}

/* scrollbar */
.chat-messages::-webkit-scrollbar {
  width: 8px;
}
.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(0,0,0,0.12);
  border-radius: 8px;
}
.chat-messages::-webkit-scrollbar-track {
  background: transparent;
}

.message {
  max-width: 70%;
  padding: 14px 18px;
  border-radius: 18px;
  line-height: 1.6;
  word-wrap: break-word;
  position: relative;
  animation: messageSlideIn 0.28s ease-out;
  font-size: 15px;
  box-sizing: border-box;
}

@keyframes messageSlideIn {
  from {
    opacity: 0;
    transform: translateY(12px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.user-message {
  align-self: flex-end;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.16);
}

.user-message::after {
  content: '';
  position: absolute;
  bottom: -6px;
  right: 18px;
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 8px solid #764ba2;
}

.ai-message {
  align-self: flex-start;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(4px);
  color: #2c3e50;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.ai-message::after {
  content: '';
  position: absolute;
  bottom: -6px;
  left: 18px;
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 8px solid rgba(255, 255, 255, 0.95);
}

.message-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.message-header b {
  font-weight: 600;
}

.tts-btn {
  padding: 6px 10px;
  border-radius: 8px;
  font-size: 12px;
  cursor: pointer;
  background: linear-gradient(135deg, #67c23a 0%, #409eff 100%);
  color: white;
  border: none;
  transition: all 0.18s ease;
  box-shadow: 0 2px 8px rgba(103, 194, 58, 0.18);
}

.tts-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(103, 194, 58, 0.25);
}

.streaming-indicator {
  color: #999;
  font-weight: 600;
  margin-left: 6px;
}

/* message content */
.message-content {
  white-space: pre-wrap;
  word-break: break-word;
}

/* input area */
.chat-input {
  padding: 24px;
  background: rgba(255, 255, 255, 0.96);
  backdrop-filter: blur(8px);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
  position: relative;
  z-index: 1;
}

.chat-input textarea {
  width: 100%;
  resize: none;
  border: 2px solid rgba(0, 0, 0, 0.06);
  border-radius: 12px;
  padding: 14px 16px;
  font-size: 15px;
  outline: none;
  background: rgba(255,255,255,0.96);
  color: #2c3e50;
  transition: all 0.18s ease;
  min-height: 20px;
  max-height: 160px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.04);
}

.chat-input textarea:focus {
  border-color: #409eff;
  box-shadow: 0 8px 30px rgba(64,158,255,0.06);
  transform: translateY(-1px);
}

.send-btn {
  position: absolute;
  right: 36px;
  bottom: 30px;
  padding: 12px 22px;
  border: none;
  border-radius: 50px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 6px 20px rgba(102,126,234,0.18);
  transition: all 0.18s ease;
}

.send-btn:hover:not(:disabled) {
  transform: translateY(-3px) scale(1.02);
}

.send-btn:disabled {
  background: #ccc;
  box-shadow: none;
  cursor: not-allowed;
}
</style>
