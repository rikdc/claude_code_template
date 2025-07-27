<template>
  <div class="conversation-detail">
    <!-- Back Navigation -->
    <div class="mb-4">
      <button @click="goBack" class="btn btn-secondary">
        <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        Back to Conversations
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadConversation" class="btn btn-primary mt-2">
        Retry
      </button>
    </div>

    <!-- Conversation Content -->
    <div v-else-if="conversation">
      <!-- Header -->
      <div class="card mb-6">
        <div class="card-header">
          <div class="flex items-start justify-between">
            <div class="flex-1 min-w-0">
              <h1 class="text-2xl font-bold text-gray-900 mb-2 break-words">
                {{ conversation.title || `Session ${conversation.session_id}` }}
              </h1>
              
              <div class="flex items-center gap-4 text-sm text-gray-500 mb-4">
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 break-all">
                  Session: {{ conversation.session_id }}
                </span>
                <span>{{ conversation.prompt_count }} messages</span>
                <span>{{ formatDate(conversation.created_at) }}</span>
                <span v-if="conversation.total_characters">
                  {{ formatCharacterCount(conversation.total_characters) }}
                </span>
              </div>
              
              <div v-if="conversation.working_directory" class="text-sm text-gray-600 font-mono mb-2 break-all">
                📁 {{ conversation.working_directory }}
              </div>
              
              <div v-if="conversation.transcript_path" class="text-sm text-gray-600 font-mono break-all">
                📄 {{ conversation.transcript_path }}
              </div>
            </div>
            
            <!-- Actions -->
            <div class="flex items-center gap-2 ml-4">
              <button @click="showRatingModal = true" class="btn btn-secondary">
                <svg class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
                Rate
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Messages -->
      <div class="space-y-4">
        <div
          v-for="(message, index) in conversation.messages"
          :key="message.id"
          class="card"
          :class="{
            'border-l-4 border-l-blue-500': message.message_type === 'prompt',
            'border-l-4 border-l-green-500': message.message_type === 'response'
          }"
        >
          <div class="card-body">
            <!-- Message Header -->
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                      :class="{
                        'bg-blue-100 text-blue-800': message.message_type === 'prompt',
                        'bg-green-100 text-green-800': message.message_type === 'response'
                      }">
                  {{ formatMessageType(message.message_type) }}
                </span>
                <span class="text-sm text-gray-500">
                  Message {{ index + 1 }}
                </span>
                <span class="text-sm text-gray-500">
                  {{ formatDate(message.timestamp) }}
                </span>
                <span v-if="message.execution_time" class="text-xs text-gray-400">
                  {{ message.execution_time }}ms
                </span>
              </div>
              
              <div class="flex items-center gap-2">
                <span class="text-xs text-gray-500">
                  {{ message.character_count }} chars
                </span>
                <button @click="copyMessage(message.content)" class="text-gray-400 hover:text-gray-600">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </button>
              </div>
            </div>
            
            <!-- Message Content -->
            <div class="prose max-w-none">
              <pre class="whitespace-pre-wrap text-sm text-gray-700 font-sans leading-relaxed break-words overflow-wrap-anywhere">{{ message.content }}</pre>
            </div>
            
            <!-- Tool Calls -->
            <div v-if="message.tool_calls && message.tool_calls.length > 0" class="mt-4 p-3 bg-gray-50 rounded-md">
              <h4 class="text-sm font-medium text-gray-700 mb-2">Tool Calls:</h4>
              <div class="space-y-2">
                <div
                  v-for="(toolCall, toolIndex) in message.tool_calls"
                  :key="toolIndex"
                  class="bg-white p-2 rounded border text-xs"
                >
                  <div class="font-medium text-gray-800">{{ toolCall.name }}</div>
                  <div v-if="toolCall.arguments" class="text-gray-600 mt-1">
                    <pre class="whitespace-pre-wrap break-words overflow-wrap-anywhere">{{ JSON.stringify(toolCall.arguments, null, 2) }}</pre>
                  </div>
                  <div v-if="toolCall.result" class="text-green-600 mt-1">
                    Result: {{ truncateText(toolCall.result, 200) }}
                  </div>
                  <div v-if="toolCall.error" class="text-red-600 mt-1">
                    Error: {{ toolCall.error }}
                  </div>
                  <div v-if="toolCall.duration" class="text-gray-500 mt-1">
                    Duration: {{ toolCall.duration }}ms
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty Messages State -->
      <div v-if="!conversation.messages || conversation.messages.length === 0" class="text-center py-8">
        <div class="text-gray-500 mb-4">
          <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-3.582 8-8 8a8.955 8.955 0 01-4.906-1.478l-3.204 1.602a1 1 0 01-1.388-1.388l1.602-3.204A8.955 8.955 0 013 12c0-4.418 3.582-8 8-8s8 3.582 8 8z" />
          </svg>
        </div>
        <h3 class="text-lg font-medium text-gray-900 mb-2">No messages found</h3>
        <p class="text-gray-500">This conversation doesn't have any messages yet.</p>
      </div>
    </div>

    <!-- Rating Modal -->
    <div v-if="showRatingModal" class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div class="mt-3">
          <h3 class="text-lg font-medium text-gray-900 mb-4">Rate this conversation</h3>
          
          <!-- Star Rating -->
          <div class="flex items-center gap-1 mb-4">
            <button
              v-for="star in 5"
              :key="star"
              @click="selectedRating = star"
              class="focus:outline-none"
            >
              <svg
                class="h-8 w-8 cursor-pointer transition-colors"
                :class="star <= selectedRating ? 'text-yellow-400' : 'text-gray-300 hover:text-yellow-300'"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
              </svg>
            </button>
          </div>
          
          <!-- Comment -->
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-2">
              Comment (optional)
            </label>
            <textarea
              v-model="ratingComment"
              rows="3"
              class="form-input"
              placeholder="What did you think about this conversation?"
            ></textarea>
          </div>
          
          <!-- Actions -->
          <div class="flex items-center justify-end gap-2">
            <button @click="closeRatingModal" class="btn btn-secondary">
              Cancel
            </button>
            <button 
              @click="submitRating" 
              :disabled="selectedRating === 0 || submittingRating"
              class="btn btn-primary"
              :class="{ 'opacity-50 cursor-not-allowed': selectedRating === 0 || submittingRating }"
            >
              {{ submittingRating ? 'Submitting...' : 'Submit Rating' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { apiService, formatters } from '../services/api.js'

export default {
  name: 'ConversationDetail',
  props: {
    id: {
      type: [String, Number],
      required: true
    }
  },
  emits: ['error'],
  setup(props, { emit }) {
    const router = useRouter()
    const route = useRoute()
    
    // Reactive state
    const conversation = ref(null)
    const loading = ref(false)
    const error = ref(null)
    const showRatingModal = ref(false)
    const selectedRating = ref(0)
    const ratingComment = ref('')
    const submittingRating = ref(false)

    // Methods
    const loadConversation = async () => {
      try {
        loading.value = true
        error.value = null
        
        const response = await apiService.getConversation(props.id)
        
        if (response.success) {
          conversation.value = response.data
        } else {
          throw new Error(response.error || 'Failed to load conversation')
        }
      } catch (err) {
        error.value = err.message
        emit('error', err)
      } finally {
        loading.value = false
      }
    }

    const goBack = () => {
      router.push('/')
    }

    const copyMessage = async (content) => {
      try {
        await navigator.clipboard.writeText(content)
        // Could add a toast notification here
        console.log('Message copied to clipboard')
      } catch (err) {
        console.error('Failed to copy message:', err)
      }
    }

    const closeRatingModal = () => {
      showRatingModal.value = false
      selectedRating.value = 0
      ratingComment.value = ''
    }

    const submitRating = async () => {
      if (selectedRating.value === 0) return
      
      try {
        submittingRating.value = true
        
        await apiService.createConversationRating(
          props.id,
          selectedRating.value,
          ratingComment.value || null
        )
        
        closeRatingModal()
        // Could show success message
        console.log('Rating submitted successfully')
        
        // Reload conversation to show updated rating
        await loadConversation()
      } catch (err) {
        console.error('Failed to submit rating:', err)
        emit('error', err)
      } finally {
        submittingRating.value = false
      }
    }

    // Utility methods
    const formatDate = formatters.formatDate
    const formatMessageType = formatters.formatMessageType
    const truncateText = formatters.truncateText
    
    const formatCharacterCount = (count) => {
      if (count < 1000) return `${count} chars`
      if (count < 10000) return `${(count / 1000).toFixed(1)}k chars`
      return `${Math.round(count / 1000)}k chars`
    }

    // Lifecycle
    onMounted(() => {
      loadConversation()
    })

    return {
      conversation,
      loading,
      error,
      showRatingModal,
      selectedRating,
      ratingComment,
      submittingRating,
      loadConversation,
      goBack,
      copyMessage,
      closeRatingModal,
      submitRating,
      formatDate,
      formatMessageType,
      truncateText,
      formatCharacterCount
    }
  }
}
</script>

<style scoped>
.conversation-detail {
  max-width: 1200px;
  margin: 0 auto;
}

.prose {
  max-width: none;
}

.prose pre {
  background: transparent;
  padding: 0;
  margin: 0;
  border: none;
  border-radius: 0;
  font-size: inherit;
  line-height: inherit;
  color: inherit;
}

.space-y-4 > :not([hidden]) ~ :not([hidden]) {
  margin-top: 1rem;
}

.space-y-2 > :not([hidden]) ~ :not([hidden]) {
  margin-top: 0.5rem;
}

.z-50 {
  z-index: 50;
}

.top-20 {
  top: 5rem;
}

.w-96 {
  width: 24rem;
}

.opacity-50 {
  opacity: 0.5;
}

.cursor-not-allowed {
  cursor: not-allowed;
}

.border-l-4 {
  border-left-width: 4px;
}

.border-l-blue-500 {
  border-left-color: #3b82f6;
}

.border-l-green-500 {
  border-left-color: #10b981;
}

/* Text wrapping utilities */
.break-all {
  word-break: break-all;
}

.break-words {
  word-break: break-word;
}

.overflow-wrap-anywhere {
  overflow-wrap: anywhere;
}

/* Ensure containers don't overflow */
.conversation-detail {
  min-width: 0; /* Allow flex items to shrink below content size */
}

.conversation-detail .flex-1 {
  min-width: 0; /* Allow flex items to shrink below content size */
}

/* Message content specific wrapping */
.prose pre {
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
  max-width: 100%;
}
</style>