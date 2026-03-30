---
inclusion: always
---

# Healing - Product Overview

Healing is a digital healthcare platform that connects patients with a broad spectrum of health specialists — human and veterinary medicine, traditional and non-traditional practices (Chinese medicine, therapies, holistic treatments, etc.). The platform centralizes each patient's medical history, enables multi-specialist collaboration, and integrates AI agents as first-class participants in the care journey.

## Vision

Every patient owns a complete, portable health record built from every consultation and treatment on the platform. Every specialist has the tools, visibility, and AI assistance to deliver better care. The platform removes friction between disciplines, enabling truly integrative health management.

## Core Domains

### 1. Specialist Management

Licensed healthcare professionals who provide care through the platform across all disciplines.

- Professional credential validation (license verification via external services)
- Detailed public profile with bio, specialty, keywords, case studies, and ratings
- Profile construction assisted by a platform-provided AI assistant
- Support for human medicine, veterinary medicine, traditional and non-traditional practices
- Email and license number uniqueness enforced
- Specialist status lifecycle: pending → active → unavailable → deleted → banned

### 2. Patient Management

Patients are the central users whose health journey the platform serves.

- Authenticated user profile with personal and contact information
- Complete medical record: all reports, documents, prescriptions, and treatment history
- Favorite specialists list for quick access
- Specialist and AI agent reviews and ratings
- Full ownership and portability of health data

### 3. Appointment Scheduling

Consultations can be booked online (video/chat) or in-person, with both human specialists and AI agents.

- Online consultations (video, chat) and physical appointment booking
- AI agent consultations: lower cost, unlimited availability per specialist
- Calendar and availability management for specialists
- Appointment reminders and notifications
- Mandatory consultation report after every appointment (human or AI)

### 4. Consultation Reports & Medical Records

Every consultation produces a structured report that feeds the patient's unified health record.

- Mandatory report generation after each consultation
- Reports linked to patient record, specialist, and appointment
- Cross-specialist report sharing: when a patient is treated by multiple specialists, reports are shared directly between them
- Document upload support (lab results, imaging, prescriptions)
- Chronological patient timeline across all specialists and disciplines

### 5. Multi-Specialist Collaboration

Patients often need care from multiple specialists. The platform enables direct collaboration.

- Shared patient context: all involved specialists see relevant reports and history
- Case discussions between specialists (structured threads per patient case)
- Referral system between specialists
- Multi-disciplinary treatment plans

### 6. AI Agents

AI is a core pillar of the platform, not an add-on.

- Each specialist can create and manage their own AI agent
- Platform provides guided instructions and tools for agent creation
- Agents are trained/fed by the specialist with their own knowledge and protocols
- Agent roles:
  - **Assistant mode**: helps the specialist during consultations, generates report drafts, suggests based on patient history
  - **Autonomous consultation mode**: patients can book and consult directly with an AI agent (with full informed consent and liability disclosure)
- AI consultations are lower cost than human consultations
- Each specialist can offer unlimited AI consultations
- Platform manages agent lifecycle, versioning, and quality monitoring

### 7. Authentication & Authorization

The platform requires authentication for all interactions.

- Mandatory login for patients and specialists
- Role-based access: patient, specialist, admin
- Secure session management
- Data access scoped by role (patients see their own records, specialists see their patients)

### 8. Reviews & Ratings

Trust and quality are built through transparent feedback.

- Patients can rate and review specialists after consultations
- Ratings visible on specialist public profiles
- Rating aggregation (average score, total reviews)
- Review moderation for inappropriate content

### 9. Search & Discovery

Patients need to find the right specialist efficiently.

- Full-text search across specialist profiles (name, specialty, keywords, description)
- Filtered search by specialty, rating, availability, practice type
- Cursor-based pagination for large result sets
- Sort by relevance, rating, recency

### 10. Notifications & Communication

Keeping all parties informed throughout the care journey.

- Appointment reminders (email, push, SMS)
- New report notifications for patients and collaborating specialists
- Case discussion notifications
- Platform announcements and onboarding guidance

## Feature Backlog

| # | Feature | Domain | Priority |
|---|---------|--------|----------|
| F1 | Specialist registration and credential validation | Specialist Management | Implemented |
| F2 | Specialist profile search with filters and pagination | Search & Discovery | Implemented |
| F3 | Patient registration and profile management | Patient Management | High |
| F4 | Authentication and authorization (patients + specialists) | Auth | High |
| F5 | Appointment scheduling (online + physical) | Scheduling | High |
| F6 | Consultation report creation and storage | Reports & Records | High |
| F7 | Patient medical record (unified timeline) | Reports & Records | High |
| F8 | Cross-specialist report sharing | Collaboration | High |
| F9 | Specialist public profile builder (with AI assistant) | Specialist Management | Medium |
| F10 | Case study publishing on specialist profile | Specialist Management | Medium |
| F11 | AI agent creation guided flow | AI Agents | Medium |
| F12 | AI agent as specialist assistant (consultation + reports) | AI Agents | Medium |
| F13 | AI agent autonomous consultations (patient-facing) | AI Agents | Medium |
| F14 | Multi-specialist case discussions | Collaboration | Medium |
| F15 | Patient favorites list | Patient Management | Medium |
| F16 | Reviews and ratings system | Reviews & Ratings | Medium |
| F17 | Notification system (email, push, SMS) | Notifications | Medium |
| F18 | Document upload and management (lab results, imaging) | Reports & Records | Medium |
| F19 | Referral system between specialists | Collaboration | Low |
| F20 | Multi-disciplinary treatment plans | Collaboration | Low |
| F21 | AI agent quality monitoring and versioning | AI Agents | Low |
| F22 | Admin dashboard (moderation, analytics, platform health) | Platform | Low |
| F23 | Payment and billing (consultations, AI consultations) | Billing | Low |
| F24 | Video/chat infrastructure for online consultations | Scheduling | Low |

## Key Business Rules

- All specialists must hold valid professional licenses (e.g., CRM, CRMV in Brazil)
- External license validation is required during onboarding
- Specialists must explicitly agree to share medical reports with patients
- Email and license number uniqueness is enforced
- Every consultation (human or AI) must produce a report — no exceptions
- AI consultations require explicit patient consent and liability acknowledgment
- Only authenticated users can access the platform
- Patients own their data and can export it at any time
- Specialists manage their own AI agents; the platform provides guardrails and monitoring
- AI agents cannot prescribe controlled substances or make definitive diagnoses without specialist review

## Architecture Philosophy

The system follows Domain-Driven Design with clear separation between:
- **Domain**: Core business logic and entities
- **Application**: Use cases and commands
- **Infrastructure**: External integrations (gRPC, databases, external APIs)

Microservice-oriented, event-driven where appropriate. Each bounded context (Specialist, Patient, Scheduling, Reports, AI Agents) is designed for independent evolution and horizontal scaling. Observability, resilience, and data sovereignty are first-class concerns.
