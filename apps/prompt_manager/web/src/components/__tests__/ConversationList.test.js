import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createRouter, createWebHistory } from 'vue-router'
import ConversationList from '../ConversationList.vue'
import * as apiService from '../../services/api.js'

// Mock the API service
vi.mock('../../services/api.js', () => ({
  apiService: {
    getConversations: vi.fn()
  },
  formatters: {
    formatRelativeTime: vi.fn((date) => '2 hours ago')
  }
}))

// Create a mock router
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: { template: 'Home' } },
    { path: '/conversation/:id', component: { template: 'Detail' } }
  ]
})

describe('ConversationList', () => {
  let wrapper
  
  const mockConversations = [
    {
      id: 1,
      session_id: 'session-123',
      title: 'Test Conversation 1',
      created_at: '2024-01-01T10:00:00Z',
      updated_at: '2024-01-01T11:00:00Z',
      prompt_count: 5,
      total_characters: 1500,
      working_directory: '/test/path',
      avg_rating: 4.5
    },
    {
      id: 2,
      session_id: 'session-456',
      title: null,
      created_at: '2024-01-02T10:00:00Z',
      updated_at: '2024-01-02T11:00:00Z',
      prompt_count: 3,
      total_characters: 800,
      working_directory: null,
      avg_rating: null
    }
  ]

  beforeEach(() => {
    // Reset mocks
    vi.clearAllMocks()
    
    // Mock successful API response
    apiService.apiService.getConversations.mockResolvedValue({
      success: true,
      data: mockConversations,
      meta: {
        total: 2,
        page: 1,
        per_page: 20
      }
    })
  })

  const createWrapper = () => {
    return mount(ConversationList, {
      global: {
        plugins: [router]
      }
    })
  }

  it('renders conversation list correctly', async () => {
    wrapper = createWrapper()
    
    // Wait for component to load
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(wrapper.find('h2').text()).toBe('Conversations')
    expect(apiService.apiService.getConversations).toHaveBeenCalledWith(1, 20)
  })

  it('displays conversations when loaded', async () => {
    wrapper = createWrapper()
    
    // Wait for API call to complete
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 0))
    
    // Check if conversations are displayed
    const conversationCards = wrapper.findAll('.card.cursor-pointer')
    expect(conversationCards).toHaveLength(2)
    
    // Check first conversation
    const firstCard = conversationCards[0]
    expect(firstCard.text()).toContain('Test Conversation 1')
    expect(firstCard.text()).toContain('session-123')
    expect(firstCard.text()).toContain('5 messages')
  })

  it('shows session ID as title when title is null', async () => {
    wrapper = createWrapper()
    
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 0))
    
    const conversationCards = wrapper.findAll('.card.cursor-pointer')
    const secondCard = conversationCards[1]
    
    expect(secondCard.text()).toContain('Session session-456')
  })

  it('handles search input changes', async () => {
    wrapper = createWrapper()
    
    const searchInput = wrapper.find('input[placeholder*="Search conversations"]')
    expect(searchInput.exists()).toBe(true)
    
    await searchInput.setValue('test query')
    expect(wrapper.vm.searchQuery).toBe('test query')
  })

  it('handles items per page change', async () => {
    wrapper = createWrapper()
    
    const perPageSelect = wrapper.find('select').at(0)
    await perPageSelect.setValue('50')
    
    expect(wrapper.vm.itemsPerPage).toBe('50')
    expect(wrapper.vm.currentPage).toBe(1) // Should reset to page 1
  })

  it('handles pagination correctly', async () => {
    // Mock response with more items to enable pagination
    apiService.apiService.getConversations.mockResolvedValue({
      success: true,
      data: mockConversations,
      meta: {
        total: 100,
        page: 1,
        per_page: 20
      }
    })
    
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(wrapper.vm.totalPages).toBe(5)
    expect(wrapper.find('.btn-primary').text()).toBe('1') // Current page button
  })

  it('shows loading state', async () => {
    // Mock a delayed response
    apiService.apiService.getConversations.mockImplementation(
      () => new Promise(resolve => setTimeout(() => resolve({
        success: true,
        data: [],
        meta: { total: 0 }
      }), 100))
    )
    
    wrapper = createWrapper()
    
    // Should show loading spinner initially
    expect(wrapper.find('.loading').exists()).toBe(true)
    expect(wrapper.find('.spinner').exists()).toBe(true)
  })

  it('shows error state when API fails', async () => {
    apiService.apiService.getConversations.mockRejectedValue(
      new Error('Failed to load conversations')
    )
    
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(wrapper.find('.error').exists()).toBe(true)
    expect(wrapper.find('.error').text()).toContain('Failed to load conversations')
  })

  it('shows empty state when no conversations', async () => {
    apiService.apiService.getConversations.mockResolvedValue({
      success: true,
      data: [],
      meta: { total: 0 }
    })
    
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 0))
    
    expect(wrapper.text()).toContain('No conversations yet')
    expect(wrapper.text()).toContain('Start using Claude Code')
  })

  it('emits conversation-count event', async () => {
    wrapper = createWrapper()
    await wrapper.vm.$nextTick()
    await new Promise(resolve => setTimeout(resolve, 0))
    
    const emitted = wrapper.emitted('conversation-count')
    expect(emitted).toBeTruthy()
    expect(emitted[0]).toEqual([2])
  })

  it('formats character count correctly', () => {
    wrapper = createWrapper()
    
    expect(wrapper.vm.formatCharacterCount(500)).toBe('500 chars')
    expect(wrapper.vm.formatCharacterCount(1500)).toBe('1.5k chars')
    expect(wrapper.vm.formatCharacterCount(15000)).toBe('15k chars')
  })
})