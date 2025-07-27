<template>
  <div id="app">
    <!-- Navigation Header -->
    <nav class="nav">
      <div class="container">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <h1 class="text-xl font-bold text-gray-900">
              <router-link to="/" class="nav-link">
                Claude Code Prompt Manager
              </router-link>
            </h1>
          </div>
          
          <div class="flex items-center gap-4">
            <router-link to="/" class="nav-link">
              Conversations
            </router-link>
            <div class="text-sm text-gray-500">
              Total: {{ totalConversations }}
            </div>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main>
      <div class="container">
        <!-- Error Display -->
        <div v-if="globalError" class="error">
          {{ globalError }}
          <button @click="clearError" class="btn btn-secondary ml-2">
            Dismiss
          </button>
        </div>

        <!-- Router View -->
        <router-view 
          @error="handleError"
          @conversation-count="updateConversationCount"
        />
      </div>
    </main>

    <!-- Footer -->
    <footer class="mt-8 py-4 text-center text-sm text-gray-500 border-t">
      <div class="container">
        Claude Code Prompt Manager - Organize and review your AI conversations
      </div>
    </footer>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { apiService } from './services/api.js'

export default {
  name: 'App',
  setup() {
    const globalError = ref(null)
    const totalConversations = ref(0)

    // Error handling
    const handleError = (error) => {
      console.error('Global error:', error)
      globalError.value = error.message || 'An unexpected error occurred'
    }

    const clearError = () => {
      globalError.value = null
    }

    // Update conversation count
    const updateConversationCount = (count) => {
      totalConversations.value = count
    }

    // Load initial data
    onMounted(async () => {
      try {
        // Get initial conversation count
        const response = await apiService.getConversations(1, 0) // Just get the first item for count
        if (response.meta && response.meta.total !== undefined) {
          totalConversations.value = response.meta.total
        }
      } catch (error) {
        console.warn('Could not load initial conversation count:', error)
        // Don't show error for this, it's not critical
      }
    })

    return {
      globalError,
      totalConversations,
      handleError,
      clearError,
      updateConversationCount
    }
  }
}
</script>

<style scoped>
/* Component-specific styles */
.nav-link {
  color: #6b7280;
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
  transition: all 0.2s ease-in-out;
}

.nav-link:hover,
.nav-link.router-link-active {
  color: #3b82f6;
  background-color: #eff6ff;
}

.text-xl {
  font-size: 1.25rem;
}

.font-bold {
  font-weight: 700;
}

.text-gray-900 {
  color: #111827;
}

.text-gray-500 {
  color: #6b7280;
}

.text-sm {
  font-size: 0.875rem;
}

.ml-2 {
  margin-left: 0.5rem;
}

.mt-8 {
  margin-top: 2rem;
}

.py-4 {
  padding-top: 1rem;
  padding-bottom: 1rem;
}

.text-center {
  text-align: center;
}

.border-t {
  border-top: 1px solid #e2e8f0;
}
</style>