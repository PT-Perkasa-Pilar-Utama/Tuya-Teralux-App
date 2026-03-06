---
apply: always
---

# Repository Rules Source

All repository AI rules are centralized in the `.agents/` directory.

Always read and apply the instructions in `.agents/` as the authoritative source of truth. Bridge files must not define parallel or conflicting rules.

Keep all future AI behavior updates centralized under `.agents/`.
