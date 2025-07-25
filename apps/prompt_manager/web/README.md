# Claude Code Prompt Manager - Frontend

A Vue.js 3 frontend application for managing and reviewing Claude Code conversations with advanced filtering, pagination, and rating capabilities.

## Features

- **Conversation List**: Browse all your Claude Code conversations with search and filtering
- **Detailed View**: View complete conversation threads with messages and tool calls
- **Rating System**: Rate conversations with 1-5 stars and add comments
- **Responsive Design**: Works seamlessly on desktop and tablet devices
- **Real-time Updates**: Dynamic loading with proper error handling and loading states
- **Modern UI**: Clean, accessible interface built with modern CSS

## Technology Stack

- **Vue.js 3** with Composition API
- **Vue Router 4** for client-side routing
- **Vite** for fast development and building
- **Axios** for API communication
- **Vitest** for unit testing
- **ESLint** for code linting
- **Prettier** for code formatting

## Prerequisites

- Node.js 18+ and npm
- Go backend service running on `http://localhost:8080`

## Quick Start

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Start development server**:
   ```bash
   npm run dev
   ```

3. **Open in browser**:
   Navigate to `http://localhost:3000`

## Available Scripts

- `npm run dev` - Start development server with hot reload
- `npm run build` - Build for production
- `npm run preview` - Preview production build locally
- `npm test` - Run unit tests
- `npm test:ui` - Run tests with UI interface
- `npm run lint` - Lint code with ESLint
- `npm run format` - Format code with Prettier

## Project Structure

```
src/
├── components/           # Vue components
│   ├── ConversationList.vue    # List view with pagination
│   └── ConversationDetail.vue  # Detailed conversation view
├── services/            # API and utility services
│   └── api.js          # Backend API communication
├── App.vue             # Root component with navigation
├── main.js             # Application entry point
└── style.css           # Global styles
```

## API Integration

The frontend communicates with the Go backend through a REST API:

- `GET /api/conversations` - List conversations with pagination
- `GET /api/conversations/:id` - Get specific conversation with messages
- `POST /api/conversations/:id/ratings` - Create conversation rating
- `GET /api/health` - Health check endpoint

### API Service Features

- **Automatic Error Handling**: User-friendly error messages
- **Request/Response Logging**: Debug API calls in development
- **Timeout Management**: 10-second timeout for all requests
- **Proxy Configuration**: Development proxy to backend on port 8080

## Component Architecture

### ConversationList.vue
- Displays paginated list of conversations
- Search functionality (ready for backend implementation)
- Sorting options by date, title, and message count
- Responsive card-based layout
- Click-through to detailed view

### ConversationDetail.vue
- Shows complete conversation with all messages
- Displays tool calls and execution details
- Message copying functionality
- Rating modal with star interface
- Back navigation to list view

### App.vue
- Global layout and navigation
- Error handling and display
- Conversation count tracking
- Responsive header with branding

## Styling Approach

- **Utility-first CSS**: Custom utility classes for rapid development
- **Component-scoped Styles**: Encapsulated styling per component
- **Responsive Design**: Mobile-first approach with breakpoints
- **Accessible Colors**: WCAG-compliant color contrast
- **Loading States**: Smooth transitions and feedback

## Development Guidelines

### Code Style
- Use Vue 3 Composition API consistently
- Reactive refs for component state
- Computed properties for derived data
- Proper error boundaries and handling

### Component Communication
- Props for parent-to-child data flow
- Events for child-to-parent communication
- Global state managed in App.vue
- API service as single source of truth

### Testing Strategy
- Unit tests for components with Vitest
- API service mocking for isolated testing
- User interaction testing with Vue Test Utils
- Component accessibility testing

## Configuration

### Development Proxy
The Vite development server proxies `/api` requests to `http://localhost:8080` to avoid CORS issues during development.

### Build Configuration
- Source maps enabled for debugging
- Optimized production builds with Vite
- Modern JavaScript output for better performance

## Deployment

1. **Build for production**:
   ```bash
   npm run build
   ```

2. **Serve static files**:
   The `dist/` directory contains all static files ready for deployment to any web server.

3. **Backend Integration**:
   Ensure the Go backend serves the frontend files and handles API routes properly.

## Future Enhancements

- [ ] Advanced search with backend integration
- [ ] Real-time updates with WebSocket
- [ ] Tag management system
- [ ] Export functionality for conversations
- [ ] Conversation statistics and analytics
- [ ] Dark mode support
- [ ] Keyboard shortcuts
- [ ] Bulk operations for conversations

## Browser Support

- Chrome/Edge 88+
- Firefox 85+
- Safari 14+
- Modern mobile browsers

## Contributing

1. Follow the existing code style and patterns
2. Add tests for new functionality
3. Update documentation for significant changes
4. Use semantic commit messages
5. Test across different screen sizes

## Troubleshooting

### Common Issues

**API Connection Issues**:
- Ensure Go backend is running on port 8080
- Check browser console for CORS errors
- Verify API endpoints match backend routes

**Build Failures**:
- Run `npm install` to ensure dependencies are up to date
- Check Node.js version compatibility
- Clear `node_modules` and reinstall if needed

**Style Issues**:
- Check for CSS class conflicts
- Verify responsive breakpoints
- Test in different browsers for compatibility

For additional help, check the browser console for error messages and ensure all dependencies are properly installed.