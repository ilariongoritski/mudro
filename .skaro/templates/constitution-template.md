# Constitution: <project name>

## Stack
- Language: <language> <version>
- Framework: <framework> <version>
- Database: <DB> <version>
- Infrastructure: <Docker/K8s/cloud>

## Coding Standards
- Linter: <name> with config <path>
- Formatter: <name>
- Naming: <conventions>
- Max function length: <N lines>
- Max nesting depth: <N levels>

## Testing
- Minimum coverage: <N%>
- Required: unit tests for business logic
- Required: integration tests for API contracts
- Framework: <name>

## Constraints
- <infrastructure constraints>
- <budget constraints>
- <compatibility requirements>

## Security
- Authorization: <model>
- Input validation: <approach>
- Secrets: <where stored>

## LLM Rules
- Do not leave stubs without explicit TODO with justification
- Do not duplicate code: prefer reuse and clear abstractions
- Do not make hidden assumptions — if unsure, ask
- Always generate AI_NOTES.md per template
- Follow the coding style described above
