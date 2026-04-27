---
name: kn-template
description: Use when generating code from templates - list, run, or create templates
---

# Working with Templates

**Announce:** "Using kn-template to work with templates."

**Core principle:** USE TEMPLATES FOR CONSISTENT CODE GENERATION.

## Inputs

- Template name or generation goal
- Variables required by prompts
- Linked pattern doc, if one exists

## ⚠️ Critical Syntax Warning

**NEVER write `$` + triple-brace in Handlebars templates:**

```handlebars
// ❌ WRONG — causes template evaluation errors
${ {{camelCase name}} }

// ✅ CORRECT — add space, use ~ for trimming
${ {{~camelCase name~}} }
```

This is the #1 cause of template generation failures. Keep this rule in mind when creating or debugging templates.

---

## Preflight

- Read the linked doc before running a non-trivial template
- Use dry run before generating real files
- Check whether a template already exists before creating a new one

## Step 1: List Templates

```json
mcp_knowns_templates({ "action": "list" })
```

## Step 2: Get Template Details

```json
mcp_knowns_templates({ "action": "get", "name": "<template-name>" })
```

Check: prompts, `doc:` link, files to generate.

## Step 3: Read Linked Documentation

```json
mcp_knowns_docs({ "action": "get", "path": "<doc-path>", "smart": true })
```

## Step 4: Run Template

```json
// Dry run first
mcp_knowns_templates({ "action": "run", "name": "<template-name>",
  "variables": { "name": "MyComponent" },
  "dryRun": true
})

// Then run for real
mcp_knowns_templates({ "action": "run", "name": "<template-name>",
  "variables": { "name": "MyComponent" },
  "dryRun": false
})
```

## Step 5: Create New Template

```json
mcp_knowns_templates({ "action": "create", "name": "<template-name>",
  "description": "Description",
  "doc": "patterns/<related-doc>"
})
```

---

## Create Template from Pattern Workflow

When `kn-extract` identifies a pattern worth templating, use this quick path:

1. **Identify**: Pattern is generalizable AND will be reused ≥3 times
2. **Check**: `mcp_knowns_templates({ "action": "list" })` — don't duplicate existing
3. **Create linked doc**: `patterns/<pattern-name>` with working examples
4. **Create template** with the linked doc:
```json
mcp_knowns_templates({ "action": "create", "name": "<pattern-name>",
  "description": "Generate <what> from pattern",
  "doc": "patterns/<pattern-name>"
})
```
5. **Validate**: `mcp_knowns_validate({ "scope": "templates" })`

**Trigger**: After `kn-extract` if pattern meets "code-generatable" criteria.

---

## Template Config Reference

```yaml
name: react-component
description: Create a React component
doc: patterns/react-component

prompts:
  - name: name
    message: Component name?
    validate: required

files:
  - template: ".tsx.hbs"
    destination: "src/components/{{kebabCase name}}.tsx"
```

---

## Versioning and Migration

If a pattern changes:

1. Update the linked `patterns/` doc first
2. Then update the template files
3. Run `mcp_knowns_validate({ "scope": "templates" })` to verify
4. Warn if existing generated files might be stale: "Pattern updated — regenerate existing files?"

---

## Validation (After Creating Template)

```json
mcp_knowns_validate({ "scope": "templates" })
```

---

## Shared Output Contract

All built-in skills in scope must end with the same user-facing information order: `kn-init`, `kn-spec`, `kn-plan`, `kn-research`, `kn-implement`, `kn-verify`, `kn-doc`, `kn-template`, `kn-extract`, and `kn-commit`.

Required order for the final user-facing response:

1. Goal/result - state what template was inspected, created, or run.
2. Key details - include which template was inspected/created/run, dry-run vs real execution, generated files, syntax warnings if applicable.
3. Next action - recommend a concrete follow-up command only when a natural handoff exists.

Keep this concise for CLI use. Template-specific content may extend the key-details section, but must not replace or reorder the shared structure.

For `kn-template`, the key details should cover:

- which template was inspected, created, or run
- dry-run vs real execution
- generated or modified files
- any missing prompt values, doc gaps, or syntax issues

When template work naturally leads to implementation or review, include the best next command. If the user only inspected templates or finished with a dry run decision, do not force a handoff.

## Failure Modes

- Missing linked doc → say so and inspect the template directly
- Dry run looks wrong → stop and fix the template before real generation
- New template overlaps an existing one → prefer update or consolidation
- **Syntax error** (`$` + triple-brace) → point to the ⚠️ Critical Syntax Warning section

## Checklist

- [ ] Listed available templates
- [ ] Read linked documentation
- [ ] Ran dry run first
- [ ] Verified generated files
- [ ] **Validated (if created new template)**
- [ ] Syntax pitfalls checked (no `$` + triple-brace)