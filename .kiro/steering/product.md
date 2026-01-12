# Product Overview

**Healing Specialist** is a healthcare platform that manages medical specialist profiles and credentials.

## Core Domain

The system focuses on onboarding and managing licensed healthcare professionals who provide specialized medical care. Key aspects:

- **Specialist Management**: Licensed healthcare professionals with verified credentials
- **Patient Empowerment**: Specialists agree to share consultation reports/medical records directly with patients
- **Professional Verification**: External license validation through third-party gateways
- **Discoverability**: Public profiles with specialties, keywords, and descriptions for patient discovery

## Business Rules

- All specialists must hold valid professional licenses (e.g., CRM in Brazil)
- Specialists must explicitly agree to share medical reports with patients during onboarding
- Email addresses and license numbers must be unique across the platform
- External license validation is required with timeout handling (800ms)

## Architecture Philosophy

The system follows Domain-Driven Design principles with clear separation between:
- **Domain Logic**: Pure business rules and entity validation
- **Application Services**: Use case orchestration with observability
- **Infrastructure**: External integrations and persistence

Dynamic aspects like availability, pricing, and scheduling are intentionally decoupled into separate services.