# Surveys Building Block - Functionality Overview

## Service Purpose

The Surveys Building Block is a microservice within the Rokwire platform that manages survey data, responses, and scoring. It provides a comprehensive system for creating, distributing, and analyzing surveys while tracking user responses and calculating scores based on configurable rules.

## Core Capabilities

### Survey Management
- **Create and manage surveys** with customizable questions and response types
- **Survey types** support various use cases (e.g., fashion quizzes, general surveys)
- **Survey metadata** including title, description, creator information, and organizational context
- **Survey lifecycle** with start/end dates, public/archived status, and completion tracking
- **Calendar integration** to link surveys to specific calendar events
- **Estimated completion time** tracking for user experience

### Survey Responses
- **Collect user responses** to survey questions with flexible data structures
- **Response tracking** with user ID, organization, and application context
- **Anonymous responses** support for sensitive surveys
- **Sensitive data handling** with configurable privacy settings
- **Response timestamps** for temporal analysis

### Scoring System
- **Configurable scoring rules** using result rules and JSON-based scoring logic
- **Score calculation** with support for multiple scoring dimensions
- **Streak tracking** for consecutive survey completions with multiplier bonuses
- **Leaderboard integration** for competitive scoring scenarios
- **Score aggregation** across multiple survey responses
- **External user ID support** for integration with external systems (Mastodon, AmgUUID)

### Analytics & Reporting
- **Survey statistics** including response counts, completion rates, and score distributions
- **Response data aggregation** for analytics and insights
- **Correct answer tracking** for quiz-style surveys
- **Score distribution analysis** with maximum scores and performance metrics

### Data Management
- **Multi-tenant support** with organization and application isolation
- **Configuration management** for environment-specific settings
- **Data deletion** with cascading cleanup logic
- **Unstructured properties** for extensible survey metadata

## API Endpoints

The service exposes REST APIs organized by role:

### Client APIs
- Get individual surveys
- List surveys with filtering (by ID, type, calendar event)
- Submit survey responses
- Retrieve user scores and rankings

### Admin APIs
- Create and update surveys
- Manage survey configurations
- View survey statistics and responses
- Configure system settings

### Analytics APIs
- Retrieve aggregated survey data
- Generate performance reports
- Access score distributions

### Building Block (BBs) APIs
- Inter-service communication
- Service-to-service data exchange

### System APIs
- Health checks and version information
- System configuration and status

## Architecture

The service follows **hexagonal architecture** with clear separation of concerns:

- **Core**: Business logic and data models
- **Driver Adapters**: REST API endpoints and HTTP handlers
- **Driven Adapters**: External dependencies (MongoDB, notifications, calendar, core services)

## External Integrations

- **MongoDB**: Primary data persistence
- **Notifications Service**: Send survey notifications to users
- **Calendar Service**: Link surveys to calendar events
- **Core Building Block**: Authentication and service registration
- **Airship**: Push notification and tagging service

## Key Features

- **Role-based access control** with admin, client, and system roles
- **CORS configuration** for cross-origin requests
- **Service account authentication** for inter-service communication
- **Configurable environment settings** stored in database
- **Comprehensive logging** for debugging and monitoring

---

## Surveys System

The service provides flexible survey functionality that allows for a variety of dynamic question types and logical flows.

### Survey Types

- **Fashion Quiz** (`fashion_quiz`): A specialized quiz type with scoring and leaderboard integration
- **General Surveys**: Flexible survey types for various use cases
- **Scored vs. Non-Scored**: Surveys can be configured as scored (with correct answers) or non-scored (opinion-based)

### Question Types

Surveys support multiple question formats through the `SurveyData` model:

- **Multiple Choice**: Single or multiple selection options with configurable correct answers
- **Numeric**: Number input with optional min/max constraints and whole number requirements
- **Text**: Free-form text responses with optional length constraints
- **DateTime**: Date and time selection with optional time component
- **Data Entry**: Structured data input with custom formats
- **Page**: Multi-question pages with navigation logic

### Question Features

- **Options with Scoring**: Each option can have an associated score value
- **Correct Answers**: Questions can define one or multiple correct answers
- **Skip Logic**: Questions can be skipped based on `allow_skip` setting
- **Follow-up Rules**: Dynamic question flow based on previous responses using `follow_up_rule`
- **Score Rules**: Custom scoring logic per question via `score_rule`
- **Maximum Scores**: Define maximum achievable score per question
- **Self-Scoring**: Optional self-assessment mode

### Survey Responses

- **Anonymous Mode**: Responses can be submitted anonymously when `anonymous: true`
- **Sensitive Data**: Special handling for sensitive surveys with `sensitive: true`
- **Response Validation**: Answers validated against correct answers and constraints
- **Completion Tracking**: Surveys track completion status and timestamps
- **Historical Data**: All responses preserved for analytics and score recalculation

## Quiz System

The service provides comprehensive quiz functionality with scoring, streaks, and competitive features.

### Scoring Mechanism

#### Basic Scoring
- Each question can contribute points based on correctness
- Scores are calculated using configurable `result_rules` and `result_json`
- Survey statistics track:
  - Total questions answered
  - Number of correct answers
  - Individual question scores
  - Maximum possible scores

#### Streak System
The service implements a daily streak system to encourage consistent participation:

- **Streak Tracking**: Monitors consecutive days of quiz completion
- **Streak Multiplier**: Scores are multiplied by **2x** when streak ≥ 2 days
- **Streak Calculation**:
  - Day 1: Normal scoring (1x multiplier)
  - Day 2+: Bonus scoring (2x multiplier)
  - Missed day: Streak resets to 0
- **Time Zones**: Uses local device time from `unstructured_properties.local_time` for accurate daily tracking
- **Streak Notifications**: Users receive notifications about their streaks

#### Score Aggregation
- **Response Count**: Total number of quizzes completed
- **Answer Count**: Total questions answered across all quizzes
- **Correct Answer Count**: Total correct answers across all quizzes
- **Cumulative Score**: Running total with streak bonuses applied
- **External ID Integration**: Scores can be linked to external systems (Mastodon, AmgUUID)

---

## Leaderboard System

The service provides a comprehensive leaderboard system for competitive quiz scenarios.

### Leaderboard Features

#### Leaderboard Management
- **Create Custom Leaderboards**: Users can create named leaderboards for specific groups
- **Multi-Leaderboard Support**: Users can participate in multiple leaderboards simultaneously
- **Admin Roles**: Leaderboard creators are automatically designated as admins
- **Member Management**: Admins can invite, remove, and manage leaderboard members

#### Leaderboard Entries
- **User Scores**: Each entry maintains a user's current score (duplicated from score record)
- **Automatic Updates**: Scores automatically sync when users complete quizzes
- **Join/Leave**: Users can join leaderboards via invitation or leave at any time
- **Admin Controls**: Admins can remove members from their leaderboards

### Ranking System

- **Real-time Rankings**: Scores are ranked in descending order
- **Rank Calculation**: Rankings update automatically when scores change
- **Tie Handling**: Users with identical scores receive the same rank
- **Pagination**: Leaderboard scores support pagination for large groups

### Leaderboard Queries

The service provides multiple ways to view leaderboard data:

- **Get Leaderboard**: Retrieve leaderboard details and metadata
- **Get Leaderboard Scores**: Paginated list of all scores in a leaderboard
- **Get User Ranks**: View a user's rank across all their leaderboards
- **Get Top and Local Scores**: Retrieve top scores plus scores near the user's position

### Notifications

The leaderboard system includes rich notification features:

#### Score Notifications
When a user completes a quiz and passes another user's score:
- Notifies users who were surpassed
- Message: "@{username} just passed you in your Runway Genius leaderboard {name}. Ready to take your spot back?"
- Includes deep link to the leaderboard

#### Daily Quiz Notifications
When a user completes their first quiz of the day in a leaderboard:
- Notifies all other leaderboard members (once per day per leaderboard)
- Message: "@{username} just scored {points} point(s) in today's Runway Genius. Can you outplay them?"
- Prevents notification spam with daily tracking via `last_quiz_time`

#### Join Notifications
When a user joins a leaderboard:
- **Admin Notification**: "@{username} accepted your invite and joined the {name} leaderboard."
- **Member Notification**: "@{username} just joined the {name} leaderboard. Want to see how they stack up?"

### Leaderboard Data Model

- **Organization/App Scoped**: Leaderboards are isolated by organization and application
- **Timestamps**: Track creation and update times
- **Last Quiz Time**: Prevents duplicate daily notifications
- **User Context**: API responses include whether the requesting user is an admin

### Transaction Safety

Leaderboard operations use database transactions to ensure consistency:
- Creating a leaderboard also creates the admin's entry atomically
- Deleting a leaderboard removes all associated entries
- Score updates sync to all leaderboard entries simultaneously