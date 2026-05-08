## ADDED Requirements

### Requirement: Distributed Lock Acquisition

The system SHALL provide distributed lock acquisition supporting lock contention based on a unique name.

#### Scenario: First acquisition succeeds

- **WHEN** a node attempts to acquire a lock that does not exist
- **THEN** the system creates the lock record and returns a lock instance

#### Scenario: Lock already held by another node

- **WHEN** a node attempts to acquire a lock that is already held and not expired
- **THEN** the system returns failure without creating a lock instance

#### Scenario: Acquire an expired lock

- **WHEN** a node attempts to acquire a lock that has expired
- **THEN** the system updates the lock record's expiration time and holder, returning a new lock instance

### Requirement: Distributed Lock Release

The system SHALL provide distributed lock release, allowing the lock holder to voluntarily release the lock.

#### Scenario: Release own lock

- **WHEN** the lock holder calls the Unlock method
- **THEN** the system sets the lock's expiration time to the current time, allowing other nodes to acquire it

#### Scenario: Release another node's lock

- **WHEN** a non-holder attempts to release the lock
- **THEN** the system takes no operation (or returns an error)

### Requirement: Lease Renewal

The system SHALL provide lease renewal, allowing the lock holder to extend the lock's validity period.

#### Scenario: Renewal succeeds

- **WHEN** the lock holder calls the Renew method before the lock expires
- **THEN** the system updates the lock's expiration time to the current time plus the lease duration

#### Scenario: Renewal fails

- **WHEN** the lock has been preempted by another node or has expired when Renew is called
- **THEN** the system returns an error indicating renewal failure

### Requirement: Lock State Check

The system SHALL provide lock state checking to determine whether the current node holds a specific lock.

#### Scenario: Check own lock

- **WHEN** the lock holder checks the lock state
- **THEN** the system returns true indicating the current node holds the lock

#### Scenario: Check another node's lock

- **WHEN** a non-holder checks the lock state
- **THEN** the system returns false indicating the current node does not hold the lock
