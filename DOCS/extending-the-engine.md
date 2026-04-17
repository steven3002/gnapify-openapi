# Extending the Engine: The Hacker's Manual

This  provides instructions for developers modifying the Conformance Engine or expanding the Validation Layer. It details the internal architecture, where specific logic resides, and how to introduce new validation rules or resolution behaviors.

## 1. Modifying the Go Conformance Engine

The CLI engine is written in standard Go, optimized for unstructured YAML/JSON traversal without relying on strict, rigid struct definitions.

### `parser.go`: OpenAPI Traversal
This file is responsible for parsing the raw OpenAPI document into memory and navigating the OpenAPI tree structure.
* **Key Logic:** The `FindOperationSecurity` function locates the target operation by traversing the `paths` object, iterating through HTTP methods, and extracting the GNAP security array. 
* **Modification Guide:** To support additional OpenAPI properties, parse custom headers, or extract extensions located outside of `components/securitySchemes/GNAP/x-gnap-access-profiles`, introduce new traversal paths here. Always utilize the `toMap` helper function to safely cast dynamically unmarshaled YAML interfaces and avoid application panics.

### `resolver.go`: Resolution and Type-Checking
This file contains the core recursive engine that merges the GNAP template with external slot data.
* **Key Logic:** The `ResolveNode` function recursively walks the profile array, identifies objects containing `$slot` keys, and maps them to the provided data payload.
* **Type-Checking Safety Net:** The `checkType` and `getExpectedType` functions handle data integrity. They intercept injected values, follow the `schemaRef` pointer to the OpenAPI component schema, and assert the runtime type against the design-time expectation.
* **Modification Guide:** To add support for advanced JSON Schema type enforcement (such as `allOf`, `anyOf`, or format validations like `uuid` or `uri`), expand the `checkType` function. 

---

## 2. Expanding the Validation Layer

The static contract is governed by Spectral. The ruleset can be modified to enforce stricter internal policies or support entirely new extension keywords.

### Custom Spectral Functions (`functions/`)
Spectral permits the creation of custom JavaScript functions for complex validation scenarios where standard JSONPath assertions are insufficient (such as cross-referencing nodes or complex string manipulations).

* **Directory:** All custom logic must reside in the `functions/` directory.
* **Creating a New Function:** Create a new JavaScript file that exports a default function. The function must accept an input target and return an array of error objects if validation fails.

```javascript
// Example: functions/customValidator.js
module.exports = (targetValue) => {
  if (targetValue.includes("illegal_character")) {
    return [
      {
        message: "The value contains an illegal character.",
      }
    ];
  }
  return;
};
```

### Registering the Function (`.spectral.yaml`)
Once the JavaScript function is authored, it must be registered and invoked within the main ruleset.

1. Declare the function under the `functions` array at the top of the `.spectral.yaml` file.
2. Create a new rule targeting a specific JSONPath array or object, and invoke the custom function in the `then` block.

```yaml
# Example: .spectral.yaml
functions:
  - customValidator

rules:
  gnap-custom-rule:
    description: "Applies the custom validation logic to all GNAP profiles."
    given: "$..x-gnap-access-profiles..*"
    severity: error
    then:
      function: customValidator
```
