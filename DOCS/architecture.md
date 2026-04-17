# System Architecture: GNAP OpenAPI MVP

This  outlines the architectural design of the GNAP Access Profiles extension for OpenAPI. 

Because the Grant Negotiation and Authorization Protocol (GNAP) relies on deeply nested, dynamic JSON arrays rather than static OAuth2 scope strings, documenting it within OpenAPI requires a unique approach. To solve this without breaking the OpenAPI 3.1.0 specification, this project is divided into two distinct, specialized subsystems: **The Validation Layer** and **The Conformance Engine**.

---

## 1. The Validation Layer (The Static Contract)

The Validation Layer is responsible for ensuring that the OpenAPI document is structurally flawless and semantically logical before it ever reaches a runtime environment or SDK. 

**Core Components:**
* `gnap-profile.schema.json` (JSON Schema Draft 7)
* `.spectral.yaml` (Spectral Ruleset)
* `functions/` (Custom JavaScript rules)

### Why Spectral and JSON Schema?
Standard JSON Schema is excellent at validating the *shape* of an isolated object (e.g., ensuring a property is a string). However, JSON Schema is fundamentally incapable of **cross-referencing** different parts of a document. 

In our architecture, an operation declares a GNAP requirement using a string pointer (e.g., `GNAP: ["incomingPaymentCreate"]`). We must verify that `"incomingPaymentCreate"` actually exists in the `components/securitySchemes/GNAP/x-gnap-access-profiles` catalog. 

We chose **Spectral** because its native JSONPath engine and custom JavaScript functions allow us to bridge this gap. Spectral acts as the "brain" of the validation layer, performing three critical tasks that standard linters cannot:
1. **Reference Integrity:** It traverses the document to guarantee that every profile requested by an operation actually exists in the catalog, preventing broken links.
2. **Orphan Detection:** It scans the catalog and warns developers if a profile is taking up space but never used by an endpoint.
3. **Slot Safety Nets:** Because JSON Schema engines (like AJV) can silently swallow errors inside recursive nodes, Spectral provides a deterministic JSONPath safety net to ensure dynamic `$slot` objects contain exactly what they need without typos.

---

## 2. The Conformance Engine (The Dynamic Resolver)

While the Validation Layer ensures the *blueprint* is valid, the Conformance Engine proves the blueprint can be transformed into a real, usable GNAP payload.

**Core Components:**
* `main.go` (CLI Router and OR-logic handler)
* `parser.go` (Unstructured YAML document traversal)
* `resolver.go` (Recursive slot injection and type-checking)

### Why Go (Golang)?
When building a tool meant to be integrated into CI/CD pipelines, API gateways, and downstream SDKs, relying on heavy runtimes (like Node.js or Python) introduces unnecessary friction and dependency bloat. 

We selected **Go** for the resolution engine for several precise reasons:
1. **Zero-Dependency Portability:** Go compiles down to a single, standalone binary executable. It can be dropped into any Alpine Linux container, GitHub Action, or local machine and run instantly.
2. **Unstructured Tree Traversal:** OpenAPI documents are massive, unpredictable YAML trees. Go's standard library and interface-handling capabilities allow the engine to cleanly unmarshal the document and traverse the nodes dynamically without requiring thousands of lines of strict struct definitions.
3. **Recursive Resolution:** GNAP templates can be infinitely nested. Go’s strong recursion performance makes it trivial to write a resolver that walks the tree, hunts for `$slot` placeholders, and injects real data.
4. **Native Type Enforcement:** The engine doesn't just blindly inject data. Go's type-assertion system intercepts the injected values, cross-references them against the OpenAPI `schemaRef` (e.g., ensuring a `WalletAddress` is actually a string), and halts execution if the types mismatch, providing an airtight safety net.

---

## System Data Flow

To understand how the pieces talk to each other, here is the linear lifecycle of a GNAP profile within this architecture:

1. **Authoring:** An API designer writes the `openapi.yaml` document, defining endpoints and the `x-gnap-access-profiles` catalog.
2. **Linting (Spectral):** The designer runs `spectral lint`. Spectral loads the JSON Schema to verify the catalog's shape, then runs custom JavaScript to verify the internal references.
3. **Providing Context:** The downstream application (or test suite) prepares a `slots.json` file containing the dynamic variables required for the specific user's session (e.g., transaction limits, wallet IDs).
4. **Execution (Go CLI):** The `gnap-profile` CLI is executed.
   * `parser.go` finds the targeted endpoint and extracts the GNAP requirement array.
   * If the endpoint allows multiple profiles (OR logic), `main.go` loops through them sequentially.
   * `resolver.go` fetches the requested profile template from the catalog.
   * `resolver.go` walks the template, injecting data from `slots.json` and verifying types.
5. **Output:** The engine emits a production-ready, spec-compliant JSON array to `stdout`, ready to be attached to a GNAP grant request.