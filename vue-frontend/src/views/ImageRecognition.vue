<template>
  <div class="image-recognition-container">
    <div class="top-bar">
      <button class="back-btn" @click="$router.push('/menu')">← 返回</button>
      <h2>AI 图像识别助手</h2>
    </div>

    <div class="chat-container" ref="chatContainerRef">
      <div
        v-for="(message, index) in messages"
        :key="index"
        :class="['message', message.role === 'user' ? 'user' : 'assistant']"
      >
        <b>{{ message.role === 'user' ? '你' : 'AI' }}:</b>
        <span>{{ message.content }}</span>
        <img v-if="message.imageUrl" :src="message.imageUrl" alt="上传的图片" />
      </div>
    </div>

    <form class="input-box" @submit.prevent="handleSubmit">
      <input
        ref="fileInputRef"
        type="file"
        accept="image/*"
        required
        @change="handleFileSelect"
      />
      <button type="submit" :disabled="!selectedFile">发送图片</button>
    </form>
  </div>
</template>

<script>
import { ref, nextTick } from 'vue'
import api from '../utils/api'

export default {
  name: 'ImageRecognition',
  setup() {
    const messages = ref([])
    const selectedFile = ref(null)
    const fileInputRef = ref()
    const chatContainerRef = ref()

    const handleFileSelect = (event) => {
      selectedFile.value = event.target.files[0]
    }

    const handleSubmit = async () => {
      if (!selectedFile.value) return

      const file = selectedFile.value
      const reader = new FileReader()

      reader.onload = async function(event) {
        const base64Url = event.target.result
        const base64Data = base64Url.split(',')[1]

        // 添加用户消息
        messages.value.push({
          role: 'user',
          content: `已上传图片: ${file.name}`,
          imageUrl: base64Url
        })

        await nextTick()
        scrollToBottom()

        try {
          const response = await api.post('/upload/send', {
            filename: file.name,
            image: base64Data
          })

          if (response.data.status_code === 1000) {
            const aiText = `识别结果: ${response.data.class_name} (置信度: ${Math.round(response.data.confidence * 100)}%)`
            messages.value.push({
              role: 'assistant',
              content: aiText
            })
          } else {
            messages.value.push({
              role: 'assistant',
              content: `[错误] ${response.data.status_msg || '识别失败'}`
            })
          }
        } catch (error) {
          console.error('Upload error:', error)
          messages.value.push({
            role: 'assistant',
            content: `[错误] 无法连接到服务器或上传失败: ${error.message}`
          })
        }

        await nextTick()
        scrollToBottom()

        // 清空文件输入
        selectedFile.value = null
        if (fileInputRef.value) {
          fileInputRef.value.value = ''
        }
      }

      reader.readAsDataURL(file)
    }

    const scrollToBottom = () => {
      if (chatContainerRef.value) {
        chatContainerRef.value.scrollTop = chatContainerRef.value.scrollHeight
      }
    }

    return {
      messages,
      selectedFile,
      fileInputRef,
      chatContainerRef,
      handleFileSelect,
      handleSubmit
    }
  }
}
</script>

<style scoped>
.image-recognition-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

.image-recognition-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="80" cy="80" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="40" cy="60" r="1" fill="rgba(255,255,255,0.1)"/><circle cx="60" cy="30" r="1.5" fill="rgba(255,255,255,0.1)"/></svg>');
  animation: float 20s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(180deg); }
}

.top-bar {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  color: white;
  display: flex;
  align-items: center;
  padding: 0 30px;
  box-shadow: 0 2px 20px rgba(0, 0, 0, 0.1);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  position: relative;
  z-index: 2;
}

.back-btn {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  padding: 10px 20px;
  border-radius: 12px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.3);
}

.top-bar h2 {
  margin: 0 0 0 20px;
  font-size: 24px;
  font-weight: 600;
  background: linear-gradient(135deg, #ffffff 0%, rgba(255,255,255,0.8) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.chat-container {
  flex: 1;
  overflow-y: auto;
  padding: 30px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  position: relative;
  z-index: 1;
}

.chat-container::-webkit-scrollbar {
  width: 6px;
}

.chat-container::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

.chat-container::-webkit-scrollbar-thumb {
  background: rgba(102, 126, 234, 0.3);
  border-radius: 3px;
}

.chat-container::-webkit-scrollbar-thumb:hover {
  background: rgba(102, 126, 234, 0.5);
}

.message {
  max-width: 70%;
  padding: 16px 20px;
  border-radius: 20px;
  line-height: 1.6;
  word-wrap: break-word;
  position: relative;
  animation: messageSlideIn 0.4s ease-out;
  font-size: 15px;
}

@keyframes messageSlideIn {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.user {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  align-self: flex-end;
  box-shadow: 0 4px 20px rgba(102, 126, 234, 0.3);
  position: relative;
}

.user::after {
  content: '';
  position: absolute;
  bottom: -2px;
  right: 20px;
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 8px solid #764ba2;
}

.assistant {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  color: #2c3e50;
  align-self: flex-start;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  position: relative;
}

.assistant::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 20px;
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 8px solid rgba(255, 255, 255, 0.95);
}

.message b {
  font-weight: 600;
  margin-right: 8px;
}

.message img {
  max-width: 250px;
  border-radius: 12px;
  display: block;
  margin-top: 12px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
  transition: all 0.3s ease;
}

.message img:hover {
  transform: scale(1.05);
}

.input-box {
  display: flex;
  padding: 30px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  gap: 20px;
  position: relative;
  z-index: 1;
}

.input-box input[type="file"] {
  flex: 1;
  border: 2px dashed #d9d9d9;
  border-radius: 12px;
  padding: 15px 20px;
  background: rgba(255, 255, 255, 0.8);
  color: #666;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 14px;
}

.input-box input[type="file"]:hover {
  border-color: #409eff;
  background: rgba(64, 158, 255, 0.05);
}

.input-box input[type="file"]::file-selector-button {
  border: none;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 8px 16px;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  font-weight: 600;
  margin-right: 12px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 10px rgba(102, 126, 234, 0.3);
}

.input-box input[type="file"]::file-selector-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}

.input-box button {
  padding: 15px 30px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.3);
}

.input-box button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
}

.input-box button:disabled {
  background: #ccc;
  box-shadow: none;
  cursor: not-allowed;
  transform: none;
}
</style>