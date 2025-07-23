## Relevant Files

- `apps/prompt_manager/cmd/main.go` - Main Go application entry point for the web service
- `apps/prompt_manager/internal/api/handlers.go` - HTTP request handlers for REST API endpoints
- `apps/prompt_manager/internal/api/handlers_test.go` - Unit tests for API handlers
- `apps/prompt_manager/internal/database/schema.sql` - SQLite database schema definition
- `apps/prompt_manager/internal/database/db.go` - Database connection and query functions
- `apps/prompt_manager/internal/database/db_test.go` - Unit tests for database operations
- `apps/prompt_manager/internal/models/conversation.go` - Data models for conversations, sessions, ratings, tags
- `apps/prompt_manager/internal/models/conversation_test.go` - Unit tests for data models
- `apps/prompt_manager/internal/processor/daemon.go` - Background queue processor for ingesting hook data
- `apps/prompt_manager/internal/processor/daemon_test.go` - Unit tests for background processor
- `apps/prompt_manager/hooks/capture-prompt.sh` - Claude Code hook script for UserPromptSubmit events
- `apps/prompt_manager/hooks/capture-completion.sh` - Claude Code hook script for Stop events
- `apps/prompt_manager/web/src/App.vue` - Main Vue.js application component
- `apps/prompt_manager/web/src/components/ConversationList.vue` - Component for displaying prompt list
- `apps/prompt_manager/web/src/components/ConversationDetail.vue` - Component for detailed conversation view
- `apps/prompt_manager/web/src/components/RatingSystem.vue` - Star rating interface component
- `apps/prompt_manager/web/src/components/TagManager.vue` - Tag input and management component
- `apps/prompt_manager/web/src/services/api.js` - Frontend API service for backend communication
- `apps/prompt_manager/web/tests/unit/components.test.js` - Unit tests for Vue components
- `.claude/settings.json` - Updated Claude Code configuration for new hooks
- `apps/prompt_manager/go.mod` - Go module dependencies
- `apps/prompt_manager/web/package.json` - Node.js/Vue.js dependencies for frontend
- `apps/prompt_manager/database/migrations/` - Directory for SQL migration scripts
- `.claude/apps/prompt_manager/database/prompts.db` - SQLite database stored in .claude path
- `.claude/apps/prompt_manager/queue/` - Directory structure for file-based queue system

### Notes

- Use `go test ./...` from `apps/prompt_manager/` to run all Go backend tests
- Use `npm test` in the `apps/prompt_manager/web/` directory to run Vue.js frontend tests
- Hook scripts must be executable and stored in `apps/prompt_manager/hooks/`
- Database stored at `.claude/apps/prompt_manager/database/prompts.db` for user data persistence
- File-based queue system uses `.claude/apps/prompt_manager/queue/` for ultra-fast hook processing
- Background daemon processes queue files asynchronously to avoid blocking Claude Code operations
- Database migrations should be versioned and support both SQLite and future PostgreSQL compatibility
- You should strive to follow Test-Driven principals were practical. Please create appropriate unit tests as you develop the system.
- Commit often, commit early.

## Tasks

- [x] 1.0 Set up Claude Code Hook Integration for Prompt Capture
  - [x] 1.1 Create directory structure for prompt manager in `apps/prompt_manager/`
  - [x] 1.2 Create queue directory structure in `.claude/apps/prompt_manager/queue/` with subdirectories (incoming, processing, failed, archive)
  - [x] 1.3 Create hook script `capture-prompt.sh` for UserPromptSubmit events with ultra-fast file writing
  - [x] 1.4 Create hook script `capture-completion.sh` for Stop events to capture Claude responses
  - [x] 1.5 Update `.claude/settings.json` to register new hooks for UserPromptSubmit and Stop events
  - [x] 1.6 Test hook integration to ensure sub-millisecond execution times
  - [x] 1.7 Implement atomic file operations with timestamp-session-event naming convention

- [ ] 2.0 Implement Backend Go Web Service with Database Layer
  - [x] 2.1 Initialize Go module and set up project structure in `apps/prompt_manager/`
  - [x] 2.2 Create SQLite database schema with conversations and ratings tables
  - [x] 2.3 Implement database connection and CRUD operations in `internal/database/`
  - [x] 2.4 Create data models for conversations, sessions, ratings, and tags in `internal/models/`
  - [x] 2.5 Design and implement REST API endpoints for conversations and ratings
  - [x] 2.6 Add health check endpoint for monitoring queue status and system health
  - [x] 2.7 Implement database migrations system for schema versioning
  - [x] 2.8 Write comprehensive unit tests for database operations and API handlers

- [ ] 3.0 Create Background Queue Processor for Data Ingestion
  - [x] 3.1 Implement file system watcher using fsnotify for queue monitoring
  - [x] 3.2 Create processing pipeline to move files between queue directories
  - [x] 3.3 Implement JSON parsing and validation for hook data
  - [x] 3.4 Build conversation correlation logic to link prompts with responses
  - [x] 3.5 Add error handling with retry logic and dead letter queue
  - [x] 3.6 Implement transcript file parsing for complete conversation context
  - [x] 3.7 Create daemon service management (systemd/launchd) configuration
  - [x] 3.8 Add comprehensive logging and monitoring for background processing

- [ ] 4.0 Create Vue.js Frontend Application with Core Components
  - [ ] 4.1 Initialize Vue.js project in `apps/prompt_manager/web/` directory
  - [ ] 4.2 Set up build configuration and development environment
  - [ ] 4.3 Create main App.vue component with routing and layout structure
  - [ ] 4.4 Build ConversationList component with filtering and pagination
  - [ ] 4.5 Implement ConversationDetail component for viewing complete conversation threads
  - [ ] 4.6 Create responsive design that works on desktop and tablet devices
  - [ ] 4.7 Implement API service layer for backend communication
  - [ ] 4.8 Add loading states, error handling, and user feedback components
  - [ ] 4.9 Write unit tests for all Vue components

- [ ] 5.0 Implement Rating and Tagging System
  - [ ] 5.1 Create RatingSystem component with 1-5 star interface
  - [ ] 5.2 Build TagManager component with autocomplete and tag suggestions
  - [ ] 5.3 Implement persistent storage of ratings and tags in database
  - [ ] 5.4 Create API endpoints for rating and tagging operations
  - [ ] 5.5 Add tag suggestion system based on previously used tags
  - [ ] 5.6 Implement optional AI tag generation using Ollama/OpenAI endpoints
  - [ ] 5.7 Create tag filtering and search functionality in conversation list
  - [ ] 5.8 Add rating trend analysis and statistics display

- [ ] 6.0 Build Session Management and Conversation Organization Features
  - [ ] 6.1 Implement automatic session grouping using Claude Code session IDs
  - [ ] 6.2 Create "Split Session" functionality for creating logical conversation breaks
  - [ ] 6.3 Implement drag-and-drop interface for reorganizing conversation blocks
  - [ ] 6.4 Build user-defined session management separate from original Claude sessions
  - [ ] 6.5 Create session merging and splitting API endpoints
  - [ ] 6.6 Add session organization persistence and history tracking
  - [ ] 6.7 Implement conversation timeline view with session boundaries
  - [ ] 6.8 Add bulk operations for managing multiple conversations
