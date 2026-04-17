# Schema Reference: x-gnap-access-profiles

This  defines the structural contract for the `x-gnap-access-profiles` OpenAPI extension. It details how to construct Grant Negotiation and Authorization Protocol (GNAP) access arrays, implement dynamic data slots, and adhere to the strict validation rules enforced by the embedded JSON Schema.

## 1. Anatomy of the GNAP Access Array

The GNAP specification requires access requests to be structured as an array of objects, rather than simple strings. To map this to OpenAPI, profiles must be defined within the document's components catalog under the GNAP security scheme.

**Location:** `components/securitySchemes/[SchemeName]/x-gnap-access-profiles`

The extension acts as a key-value map. The keys serve as the reference names (pointers) used by operation endpoints. The values are the strict GNAP access arrays.

```yaml
components:
  securitySchemes:
    GNAP:
      type: http
      scheme: GNAP
      x-gnap-access-profiles:
        exampleProfileName:
          - type: "incoming-payment"
            actions: ["create", "read"]
```

## 2. Static vs. Dynamic Profiles

Profiles are categorized into two types based on their data requirements:

* **Static Profiles:** These contain fully hardcoded values. They represent constant access requirements that never change between requests. A static profile can be submitted to an Authorization Server exactly as written.
* **Dynamic Profiles:** These contain placeholder objects that must be populated with runtime or session-specific data (such as transaction limits, account identifiers, or user IDs) before the grant is valid.

## 3. The `$slot` Syntax

To create a dynamic profile, primitive values (strings, numbers, booleans) within the access array are replaced with a Typed Slot Placeholder object. The conformance engine uses this syntax to inject external data.

A valid slot object must contain specific keys:
* `$slot` (String): The exact lookup key matching the external data source.
* `schemaRef` (String): A local OpenAPI JSON pointer (e.g., `#/components/schemas/Amount`) defining the expected data type.
* `required` (Boolean, Optional): Indicates if the resolution should fail if the data is missing. Defaults to `true`.

**Example of a Dynamic Slot:**
```yaml
outgoingPaymentCreate:
  - type: "outgoing-payment"
    actions: ["create"]
    limits:
      debitAmount:
        $slot: "limits.debitAmount"
        schemaRef: "#/components/schemas/PaymentAmount"
        required: true
```

## 4. Strict Contract Rules

The provided JSON Schema (`gnap-profile.schema.json`) enforces strict structural integrity to prevent malformed grants. The following constraints are automatically validated:

### The `oneOf` Node Logic
To allow dynamic data injection without breaking schema validation, the schema utilizes strict `oneOf` logic for all property values. A node within the access array is permitted to be:
* A string
* A number
* A boolean
* An array of valid nodes
* A standard object (which strictly forbids `$slot` or `schemaRef` keys at its root level)
* A Typed Slot Placeholder

This constraint prevents hybrid objects that attempt to mix standard properties with slot directives.

### The `$slot` and `schemaRef` Pairing
If an object is declared as a Typed Slot Placeholder, the schema and internal Spectral rules enforce a strict dependency constraint. The object must contain exactly both `$slot` and `schemaRef`. If one is present without the other, or if a typographical error occurs (e.g., `$slots` instead of `$slot`), the validation layer automatically rejects the document. This mechanism ensures the conformance engine always has the necessary type-checking context to resolve the data safely.
