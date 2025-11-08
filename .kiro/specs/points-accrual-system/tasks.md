# Implementation Plan

- [x] 1. Create database schema and migrations for balance and withdrawal tables
  - Create migration for user_balances table with current and withdrawn fields
  - Create migration for withdrawals table with order_number, amount, and processed_at fields
  - Add necessary indexes for performance optimization
  - _Requirements: 5.4_

- [x] 2. Implement balance management domain
- [x] 2.1 Create balance models and repository interface
  - Define Balance struct with JSON tags for API responses
  - Create BalanceRepository interface with CRUD operations
  - Implement DatabaseRepository for PostgreSQL with transaction support
  - _Requirements: 1.1, 5.1, 5.3_

- [x] 2.2 Implement balance service with business logic
  - Create BalanceService with methods for getting and updating balances
  - Implement atomic balance operations using database transactions
  - Add validation and error handling for balance operations
  - _Requirements: 1.1, 5.1, 5.2_

- [x] 2.3 Create balance HTTP handler and routes
  - Implement GetBalance handler for GET /api/user/balance endpoint
  - Add proper JSON response formatting and error handling
  - Register route in main.go with authentication middleware
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 3. Implement withdrawal management domain
- [x] 3.1 Create withdrawal models and repository
  - Define Withdrawal struct with proper JSON tags
  - Create WithdrawalRepository interface with database operations
  - Implement DatabaseRepository with sorting by processed_at DESC
  - _Requirements: 2.1, 3.1_

- [x] 3.2 Implement withdrawal service with validation
  - Create WithdrawalService with balance validation logic
  - Implement Luhn algorithm validation for order numbers
  - Add insufficient funds checking and atomic balance deduction
  - _Requirements: 2.1, 2.2, 2.3, 5.2_

- [x] 3.3 Create withdrawal HTTP handlers and routes
  - Implement WithdrawPoints handler for POST /api/user/balance/withdraw
  - Implement GetWithdrawals handler for GET /api/user/withdrawals
  - Add proper error responses (402, 422) and JSON formatting
  - Register routes in main.go with authentication middleware
  - _Requirements: 2.1, 2.2, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4_

- [x] 4. Implement accrual system integration
- [x] 4.1 Create accrual client for external API communication
  - Implement AccrualClient struct with HTTP client configuration
  - Add GetOrderAccrual method with proper error handling
  - Implement retry logic for 429 responses with Retry-After header
  - Add timeout and exponential backoff for failed requests
  - _Requirements: 4.4, 4.5_

- [x] 4.2 Create accrual response models and validation
  - Define AccrualResponse struct matching external API format
  - Add validation for response status values (REGISTERED, PROCESSING, INVALID, PROCESSED)
  - Handle optional accrual field for different response types
  - _Requirements: 4.2, 4.3_

- [x] 4.3 Implement background worker for order processing
  - Create AccrualWorker struct with configurable polling interval
  - Implement processOrders method to fetch and update order statuses
  - Add logic to query orders with NEW and PROCESSING statuses
  - Implement status updates and balance additions for PROCESSED orders
  - _Requirements: 4.1, 4.2, 4.3_

- [x] 5. Integrate accrual worker with main application
- [x] 5.1 Update order repository with accrual field support
  - Modify existing Order model to include accrual field in JSON responses
  - Update GetOrdersByUserID to return accrual information
  - Ensure existing order functionality remains unchanged
  - _Requirements: 4.2_

- [x] 5.2 Initialize and start accrual worker in main.go
  - Create AccrualWorker instance with proper dependencies
  - Start worker as goroutine with context cancellation
  - Add graceful shutdown handling for worker
  - _Requirements: 4.1, 6.1, 6.2, 6.3_

- [x] 5.3 Add automatic balance creation for new users
  - Modify user registration to create initial balance record
  - Ensure balance exists before any balance operations
  - Add migration to create balances for existing users
  - _Requirements: 5.4_

- [x] 6. Add comprehensive error handling and logging
  - Add structured logging for all balance and withdrawal operations
  - Implement proper error responses for all edge cases
  - Add monitoring for external service availability
  - Log all accrual worker activities and errors
  - _Requirements: 1.3, 2.5, 3.4, 4.5, 6.3_

- [x] 7. Create unit tests for all components
  - Write tests for BalanceService with mock repository
  - Write tests for WithdrawalService including validation scenarios
  - Write tests for AccrualClient with mock HTTP responses
  - Write tests for AccrualWorker with mock dependencies
  - _Requirements: All requirements_

- [x] 8. Create integration tests for API endpoints
  - Test complete balance retrieval flow with database
  - Test withdrawal flow including insufficient funds scenarios
  - Test withdrawal history retrieval with proper sorting
  - Test error responses for unauthorized and invalid requests
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4_