<!--
  SYNC IMPACT REPORT
  ==================
  Version Change: Initial → 1.0.0
  Constitution Type: MINOR (Initial establishment)
  
  Principles Established:
  - I. Code Quality First (NEW)
  - II. Testing Standards (NEW)
  - III. User Experience Consistency (NEW)
  - IV. Performance Requirements (NEW)
  
  Template Validation Status:
  ✅ plan-template.md - Constitution Check section compatible
  ✅ spec-template.md - Requirements alignment verified
  ✅ tasks-template.md - Task categorization aligns with principles
  ✅ agent-file-template.md - No updates required
  ✅ checklist-template.md - No updates required
  
  Follow-up Items: None
-->

# Wink-Code Constitution

## Core Principles

### I. Code Quality First

**Declaration**: All code contributions MUST meet stringent quality standards before merge.

**Non-Negotiable Rules**:
- Code MUST pass linting and formatting checks (no warnings tolerated in CI)
- Type hints/annotations MUST be present for all public interfaces
- Complexity metrics MUST remain below thresholds: cyclomatic complexity ≤10 per function
- Code duplication MUST NOT exceed 3% similarity across the codebase
- All public functions/classes MUST have clear, complete documentation strings
- Code reviews MUST verify readability, maintainability, and adherence to project patterns

**Rationale**: As a CLI coding agent, wink-code directly impacts developer workflows. Poor code 
quality creates maintenance burden, reduces trust, and undermines the tool's value proposition. 
High-quality code ensures the project remains maintainable, extensible, and reliable as it scales.

### II. Testing Standards (NON-NEGOTIABLE)

**Declaration**: Comprehensive testing is mandatory for all features and bug fixes.

**Non-Negotiable Rules**:
- Test-Driven Development (TDD) MUST be followed: Write tests → Verify they fail → Implement → Verify pass
- Unit test coverage MUST be ≥90% for all new code (measured by line and branch coverage)
- Integration tests MUST cover all API contract changes and inter-component communication
- Edge cases and error paths MUST have explicit test coverage
- Tests MUST be deterministic, isolated, and fast (unit tests <100ms each)
- Flaky tests are not acceptable; intermittent failures MUST be fixed immediately
- Breaking test changes require explicit documentation and version bump justification

**Rationale**: Testing ensures reliability and prevents regressions. Given wink-code's role in code 
generation and LLM integration, rigorous testing prevents cascading failures that could corrupt user 
projects or generate incorrect code. TDD enforces design quality and provides living documentation.

### III. User Experience Consistency

**Declaration**: User interactions MUST be predictable, intuitive, and consistent across all features.

**Non-Negotiable Rules**:
- CLI commands MUST follow consistent naming conventions and argument patterns
- Error messages MUST be actionable, specific, and user-friendly (no raw stack traces to users)
- Output formats MUST be consistent: JSON for machine-readable, human-readable for direct consumption
- stdin/stdout/stderr protocols MUST be respected: data on stdout, errors on stderr, no mixing
- Response times for interactive commands MUST be ≤2 seconds for feedback initiation
- Documentation MUST include examples for every user-facing command and common workflows
- Breaking UX changes require migration guides and deprecation warnings (minimum 1 version cycle)

**Rationale**: Consistency reduces cognitive load and learning curve. CLI tools are used repetitively; 
inconsistent behavior frustrates users and reduces productivity. Clear error messages and predictable 
patterns enable users to self-serve and build automation confidently.

### IV. Performance Requirements

**Declaration**: Performance MUST be optimized to respect user time and system resources.

**Non-Negotiable Rules**:
- LLM API calls MUST implement timeout limits (default: 30s) and retry logic with exponential backoff
- File operations MUST be efficient: streaming for large files, batching for multiple operations
- Memory footprint MUST remain ≤500MB for typical workloads (excluding LLM provider memory)
- CLI startup time MUST be ≤500ms (cold start, excluding network requests)
- Token consumption MUST be monitored and optimized; unnecessary context bloat is not acceptable
- Performance regressions >10% require explicit justification before merge
- Benchmarks MUST be maintained for critical paths: code generation, file parsing, API interactions

**Rationale**: Performance directly impacts developer experience. Slow tools disrupt flow state and 
reduce adoption. Given wink-code's integration with potentially slow LLM APIs, local operations must 
be highly optimized. Resource efficiency ensures the tool runs reliably in constrained environments.

## Development Workflow

**Code Review Requirements**:
- All changes MUST go through pull request review (minimum 1 approver)
- Reviewer MUST verify constitution compliance explicitly
- Security-sensitive changes require additional security review
- Breaking changes require maintainer approval and version bump discussion

**Quality Gates**:
- CI pipeline MUST pass all checks: linting, type checking, tests, coverage thresholds
- Documentation MUST be updated for user-facing changes
- Changelog MUST be updated with user-impact description

**Testing Gates**:
- Unit tests MUST pass with ≥90% coverage
- Integration tests MUST pass for affected components
- Performance benchmarks MUST not regress >10% without justification

## Technology Standards

**Language & Dependencies**:
- Primary language: Python 3.11+ (for modern type hints and performance)
- Dependency management: Use established, maintained libraries; minimize dependency count
- Security: Dependencies MUST be scanned for vulnerabilities; critical CVEs block merge

**Compatibility**:
- MUST support Windows, macOS, Linux (cross-platform compatibility non-negotiable)
- API compatibility with OpenAI-compatible endpoints MUST be maintained

## Governance

**Constitutional Authority**: This constitution supersedes all other development practices and 
decisions. When conflicts arise, constitution principles take precedence.

**Amendment Process**:
- Amendments require documented rationale and impact analysis
- Breaking principle changes require team consensus and migration plan
- Version bumps follow semantic versioning:
  - MAJOR: Backward-incompatible governance changes, principle removals/redefinitions
  - MINOR: New principles added, material expansions to existing guidance
  - PATCH: Clarifications, wording refinements, non-semantic updates

**Compliance & Enforcement**:
- All pull requests MUST include constitution compliance verification
- Violations MUST be documented and justified or rejected
- Complexity additions MUST be explicitly justified with rationale
- Regular constitution reviews (quarterly) to ensure relevance and effectiveness

**Complexity Justification**: Any violation or exception to these principles MUST be documented with:
- Clear technical/business rationale
- Risk assessment and mitigation plan
- Approval from project maintainers
- Plan to remediate or conform in future iterations if applicable

**Version**: 1.0.0 | **Ratified**: 2025-11-10 | **Last Amended**: 2025-11-10
