<template>
  <div class="conversation-list">
    <!-- Header with Search and Filters -->
    <div class="card mb-4">
      <div class="card-header">
        <div class="flex items-center justify-between">
          <h2>Conversations</h2>
          <div class="text-sm text-gray-500">
            {{ conversations.length }} of {{ totalConversations }} conversations
          </div>
        </div>
      </div>
      
      <div class="card-body">
        <div class="flex flex-col gap-4 md:flex-row md:items-center">
          <!-- Search Input -->
          <div class="flex-1">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search conversations by title or session ID..."
              class="form-input"
              @input="handleSearch"
            />
          </div>
          
          <!-- Items per page -->
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600">Show:</label>
            <select 
              v-model="itemsPerPage" 
              @change="handleItemsPerPageChange"
              class="form-input w-20"
            >
              <option value="10">10</option>
              <option value="20">20</option>
              <option value="50">50</option>
            </select>
          </div>
          
          <!-- Sort Options -->
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600">Sort:</label>
            <select 
              v-model="sortBy" 
              @change="handleSortChange"
              class="form-input w-32"
            >
              <option value="updated_at">Recent</option>
              <option value="created_at">Created</option>
              <option value="title">Title</option>
              <option value="prompt_count">Messages</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadConversations" class="btn btn-primary mt-2">
        Retry
      </button>
    </div>

    <!-- Empty State -->
    <div v-else-if="conversations.length === 0" class="text-center py-8">
      <div class="text-gray-500 mb-4">
        <svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-3.582 8-8 8a8.955 8.955 0 01-4.906-1.478l-3.204 1.602a1 1 0 01-1.388-1.388l1.602-3.204A8.955 8.955 0 013 12c0-4.418 3.582-8 8-8s8 3.582 8 8z" />
        </svg>
      </div>
      <h3 class="text-lg font-medium text-gray-900 mb-2">
        {{ searchQuery ? 'No matching conversations' : 'No conversations yet' }}
      </h3>
      <p class="text-gray-500">
        {{ searchQuery ? 'Try adjusting your search terms' : 'Start using Claude Code to see your conversations here' }}
      </p>
    </div>

    <!-- Conversation Cards -->
    <div v-else class="space-y-4">
      <div
        v-for="conversation in conversations"
        :key="conversation.id"
        class="card cursor-pointer transition-all hover:shadow-md"
        @click="viewConversation(conversation.id)"
      >
        <div class="card-body">
          <div class="flex items-start justify-between">
            <div class="flex-1 min-w-0">
              <!-- Title and Session -->
              <div class="flex items-center gap-2 mb-2">
                <h3 class="text-lg font-medium text-gray-900 truncate">
                  {{ conversation.title || `Session ${conversation.session_id}` }}
                </h3>
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                  {{ conversation.session_id }}
                </span>
              </div>
              
              <!-- Metadata -->
              <div class="flex items-center gap-4 text-sm text-gray-500 mb-2">
                <span class="flex items-center gap-1">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-3.582 8-8 8a8.955 8.955 0 01-4.906-1.478l-3.204 1.602a1 1 0 01-1.388-1.388l1.602-3.204A8.955 8.955 0 013 12c0-4.418 3.582 8-8 8z" />
                  </svg>
                  {{ conversation.prompt_count }} messages
                </span>
                <span class="flex items-center gap-1">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {{ formatRelativeTime(conversation.updated_at) }}
                </span>
                <span v-if="conversation.total_characters" class="flex items-center gap-1">
                  <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  {{ formatCharacterCount(conversation.total_characters) }}
                </span>
              </div>
              
              <!-- Working Directory -->
              <div v-if="conversation.working_directory" class="text-xs text-gray-400 font-mono truncate">
                📁 {{ conversation.working_directory }}
              </div>
            </div>
            
            <!-- Actions -->
            <div class="flex items-center gap-2 ml-4">
              <!-- Rating Display -->
              <div v-if="conversation.avg_rating" class="flex items-center gap-1">
                <div class="flex">
                  <svg
                    v-for="star in 5"
                    :key="star"
                    class="h-4 w-4"
                    :class="star <= Math.round(conversation.avg_rating) ? 'text-yellow-400' : 'text-gray-300'"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                  </svg>
                </div>
                <span class="text-xs text-gray-500">{{ conversation.avg_rating.toFixed(1) }}</span>
              </div>
              
              <!-- View button -->
              <svg class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="!loading && conversations.length > 0" class="mt-6 flex items-center justify-between">
      <div class="text-sm text-gray-700">
        Showing {{ ((currentPage - 1) * itemsPerPage) + 1 }} to 
        {{ Math.min(currentPage * itemsPerPage, totalConversations) }} of 
        {{ totalConversations }} results
      </div>
      
      <div class="flex items-center gap-2">
        <button
          @click="previousPage"
          :disabled="currentPage <= 1"
          class="btn btn-secondary"
          :class="{ 'opacity-50 cursor-not-allowed': currentPage <= 1 }"
        >
          Previous
        </button>
        
        <div class="flex items-center gap-1">
          <button
            v-for="page in visiblePages"
            :key="page"
            @click="goToPage(page)"
            class="btn"
            :class="page === currentPage ? 'btn-primary' : 'btn-secondary'"
          >
            {{ page }}
          </button>
        </div>
        
        <button
          @click="nextPage"
          :disabled="currentPage >= totalPages"
          class="btn btn-secondary"
          :class="{ 'opacity-50 cursor-not-allowed': currentPage >= totalPages }"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { apiService, formatters } from '../services/api.js'

export default {
  name: 'ConversationList',
  emits: ['error', 'conversation-count'],
  setup(props, { emit }) {
    const router = useRouter()
    
    // Reactive state
    const conversations = ref([])
    const loading = ref(false)
    const error = ref(null)
    const searchQuery = ref('')
    const currentPage = ref(1)
    const itemsPerPage = ref(20)
    const totalConversations = ref(0)
    const sortBy = ref('updated_at')
    
    // Computed properties
    const totalPages = computed(() => Math.ceil(totalConversations.value / itemsPerPage.value))
    
    const visiblePages = computed(() => {
      const pages = []
      const start = Math.max(1, currentPage.value - 2)
      const end = Math.min(totalPages.value, currentPage.value + 2)
      
      for (let i = start; i <= end; i++) {
        pages.push(i)
      }
      return pages
    })

    // Methods
    const loadConversations = async () => {
      try {
        loading.value = true
        error.value = null
        
        const response = await apiService.getConversations(currentPage.value, itemsPerPage.value)
        
        if (response.success) {
          conversations.value = response.data || []
          totalConversations.value = response.meta?.total || 0
          emit('conversation-count', totalConversations.value)
        } else {
          throw new Error(response.error || 'Failed to load conversations')
        }
      } catch (err) {
        error.value = err.message
        emit('error', err)
      } finally {
        loading.value = false
      }
    }

    const viewConversation = (id) => {
      router.push(`/conversation/${id}`)
    }

    const handleSearch = () => {
      // Reset to first page when searching
      currentPage.value = 1
      // TODO: Implement search functionality when backend supports it
      console.log('Search:', searchQuery.value)
    }

    const handleItemsPerPageChange = () => {
      currentPage.value = 1
      loadConversations()
    }

    const handleSortChange = () => {
      currentPage.value = 1
      // TODO: Implement sorting when backend supports it
      console.log('Sort by:', sortBy.value)
      loadConversations()
    }

    const previousPage = () => {
      if (currentPage.value > 1) {
        currentPage.value--
        loadConversations()
      }
    }

    const nextPage = () => {
      if (currentPage.value < totalPages.value) {
        currentPage.value++
        loadConversations()
      }
    }

    const goToPage = (page) => {
      currentPage.value = page
      loadConversations()
    }

    // Utility methods
    const formatRelativeTime = formatters.formatRelativeTime
    
    const formatCharacterCount = (count) => {
      if (count < 1000) return `${count} chars`
      if (count < 10000) return `${(count / 1000).toFixed(1)}k chars`
      return `${Math.round(count / 1000)}k chars`
    }

    // Watchers
    watch(currentPage, loadConversations)

    // Lifecycle
    onMounted(() => {
      loadConversations()
    })

    return {
      conversations,
      loading,
      error,
      searchQuery,
      currentPage,
      itemsPerPage,
      totalConversations,
      sortBy,
      totalPages,
      visiblePages,
      loadConversations,
      viewConversation,
      handleSearch,
      handleItemsPerPageChange,
      handleSortChange,
      previousPage,
      nextPage,
      goToPage,
      formatRelativeTime,
      formatCharacterCount
    }
  }
}
</script>

<style scoped>
.conversation-list {
  max-width: 1200px;
  margin: 0 auto;
}

.space-y-4 > :not([hidden]) ~ :not([hidden]) {
  margin-top: 1rem;
}

.opacity-50 {
  opacity: 0.5;
}

.cursor-not-allowed {
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .md\:flex-row {
    flex-direction: column;
  }
  
  .md\:items-center {
    align-items: stretch;
  }
}
</style>