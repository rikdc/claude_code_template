import axios from 'axios'

// Create axios instance with base configuration
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor for logging
api.interceptors.request.use(
  (config) => {
    console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`)
    return config
  },
  (error) => {
    console.error('API Request Error:', error)
    return Promise.reject(error)
  }
)

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => {
    console.log(`API Response: ${response.status} ${response.config.url}`)
    return response
  },
  (error) => {
    console.error('API Response Error:', error.response?.data || error.message)
    
    // Handle common error scenarios
    if (error.response?.status === 404) {
      throw new Error('Resource not found')
    } else if (error.response?.status === 500) {
      throw new Error('Server error - please try again later')
    } else if (error.code === 'ECONNABORTED') {
      throw new Error('Request timeout - please check your connection')
    } else if (error.response?.data?.error) {
      throw new Error(error.response.data.error)
    } else {
      throw new Error(error.message || 'An unexpected error occurred')
    }
  }
)

// API service functions
export const apiService = {
  // Health check
  async healthCheck() {
    const response = await api.get('/health')
    return response.data
  },

  // Conversations
  async getConversations(page = 1, perPage = 20) {
    const response = await api.get('/conversations', {
      params: { page, per_page: perPage }
    })
    return response.data
  },

  async getConversation(id) {
    const response = await api.get(`/conversations/${id}`)
    return response.data
  },

  async createConversation(conversationData) {
    const response = await api.post('/conversations', conversationData)
    return response.data
  },

  async updateConversation(id, conversationData) {
    const response = await api.put(`/conversations/${id}`, conversationData)
    return response.data
  },

  async deleteConversation(id) {
    const response = await api.delete(`/conversations/${id}`)
    return response.data
  },

  // Ratings
  async createConversationRating(conversationId, rating, comment = null) {
    const response = await api.post(`/conversations/${conversationId}/ratings`, {
      rating,
      comment
    })
    return response.data
  },

  async getConversationRatings(conversationId) {
    const response = await api.get(`/conversations/${conversationId}/ratings`)
    return response.data
  },

  async updateRating(ratingId, rating, comment = null) {
    const response = await api.put(`/ratings/${ratingId}`, {
      rating,
      comment
    })
    return response.data
  },

  async deleteRating(ratingId) {
    const response = await api.delete(`/ratings/${ratingId}`)
    return response.data
  },

  async getRatingStats() {
    const response = await api.get('/ratings/stats')
    return response.data
  }
}

// Utility functions for data formatting
export const formatters = {
  // Format date for display
  formatDate(dateString) {
    const date = new Date(dateString)
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { 
      hour: '2-digit', 
      minute: '2-digit' 
    })
  },

  // Format relative time (e.g., "2 hours ago")
  formatRelativeTime(dateString) {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now - date
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins} minute${diffMins === 1 ? '' : 's'} ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours === 1 ? '' : 's'} ago`
    if (diffDays < 7) return `${diffDays} day${diffDays === 1 ? '' : 's'} ago`
    
    return this.formatDate(dateString)
  },

  // Truncate text to specified length
  truncateText(text, maxLength = 100) {
    if (!text) return ''
    if (text.length <= maxLength) return text
    return text.substring(0, maxLength) + '...'
  },

  // Format message type for display
  formatMessageType(type) {
    return type === 'prompt' ? 'User' : 'Claude'
  },

  // Generate rating stars
  generateStars(rating, maxRating = 5) {
    const stars = []
    for (let i = 1; i <= maxRating; i++) {
      stars.push({
        filled: i <= rating,
        value: i
      })
    }
    return stars
  }
}

// Constants for the application
export const constants = {
  MESSAGE_TYPES: {
    PROMPT: 'prompt',
    RESPONSE: 'response'
  },
  
  RATING_RANGE: {
    MIN: 1,
    MAX: 5
  },

  PAGINATION: {
    DEFAULT_PAGE_SIZE: 20,
    MAX_PAGE_SIZE: 100
  }
}

export default apiService