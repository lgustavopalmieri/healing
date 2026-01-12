---
inclusion: always
---

# Healing Specialist - Product Overview

A healthcare platform service focused on specialist onboarding and management. The system enables medical specialists to register, validate their credentials, and create discoverable profiles for patients.

## Core Domain

**Specialist Management**: Licensed healthcare professionals who provide specialized medical care through the platform. Key features include:

- Professional credential validation (license verification)
- Patient data sharing agreements (specialists agree to share consultation reports directly with patients)
- Public profile creation for patient discovery
- Specialty-based categorization and keyword tagging

## Key Business Rules

- All specialists must hold valid professional licenses (e.g., CRM in Brazil)
- External license validation is required during onboarding
- Specialists must explicitly agree to share medical reports with patients
- Email and license number uniqueness is enforced
- Dynamic aspects (availability, pricing, scheduling) are handled by separate services

## Architecture Philosophy

The system follows Domain-Driven Design with clear separation between:
- **Domain**: Core business logic and entities
- **Application**: Use cases and commands
- **Infrastructure**: External integrations (gRPC, databases, external APIs)

Focus on immutable identity and professional credentials, with decoupled services for dynamic operations.