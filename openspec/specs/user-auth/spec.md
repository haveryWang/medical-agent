# User Auth Specification

## Purpose
Define user login, session enforcement, role-aware navigation, and logout behavior.

## Requirements

### Requirement: User login
The system SHALL allow an active user to log in with an account identifier and password.

#### Scenario: Successful login
- **WHEN** an active user submits valid credentials
- **THEN** the system returns an authenticated session token and basic user profile data

#### Scenario: Invalid login
- **WHEN** a user submits an unknown account or incorrect password
- **THEN** the system rejects the request with a generic authentication error

### Requirement: Session enforcement
The system SHALL require an authenticated session for all non-login application APIs.

#### Scenario: Missing session
- **WHEN** a request without a valid session token calls a protected API
- **THEN** the system rejects the request with an unauthorized response

#### Scenario: Expired session
- **WHEN** a request uses an expired session token
- **THEN** the system rejects the request and requires the frontend to return to login

### Requirement: Role-aware navigation
The system SHALL expose user role and permission data so the frontend can show permitted management areas.

#### Scenario: Knowledge manager navigation
- **WHEN** a user with knowledge-base permissions logs in
- **THEN** the frontend displays knowledge-base list, document upload, tag management, scenario management, and department management navigation items

#### Scenario: Restricted user navigation
- **WHEN** a user lacks management permissions
- **THEN** the frontend hides management-only navigation items and backend APIs still enforce access control

### Requirement: Logout
The system SHALL allow authenticated users to terminate the current session.

#### Scenario: User logs out
- **WHEN** an authenticated user triggers logout
- **THEN** the backend invalidates the session and the frontend clears local authentication state
