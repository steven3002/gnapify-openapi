# CLI Usage Guide: The Conformance Engine

This  provides a comprehensive guide to executing the `gnap-profile` command-line interface (CLI). The compiled Go binary serves as the Conformance Engine, responsible for parsing OpenAPI specifications, locating GNAP security profiles, injecting dynamic data, and outputting compliant JSON grant arrays.

## 1. Command Flags

The CLI accepts specific flags to control the resolution process. 

### `--spec` (Required)
* **Description:** The relative or absolute file path to the OpenAPI specification document.
* **Expected Format:** `.yaml` or `.json`
* **Default:** None (Execution halts if omitted)

### `--operationId` (Required)
* **Description:** The exact `operationId` defined in the OpenAPI paths object that requires GNAP resolution.
* **Expected Format:** String
* **Default:** None (Execution halts if omitted)

### `--slots` (Optional)
* **Description:** The file path to a JSON document containing dynamic key-value pairs. These values are used to populate matching `$slot` placeholders defined within the target GNAP profile.
* **Expected Format:** `.json`
* **Default:** None. (Note: If omitted when attempting to resolve a profile that contains `required: true` slots, execution will fail).

### `--output` (Optional)
* **Description:** Determines the formatting of the standard output upon successful resolution.
* **Expected Format:** String (`text` or `json`)
* **Default:** `text`

---

## 2. Exit Codes

The engine strictly adheres to standard POSIX exit codes to ensure seamless integration into CI/CD pipelines, automated test suites, and downstream SDKs.

* **`Exit Code 0` (Success):** The OpenAPI document was parsed, the operation was found, all necessary slots were successfully injected and type-checked, and the final GNAP grant was printed to standard output (`stdout`).
* **`Exit Code 1` (Failure):** The execution failed. This occurs due to missing required flags, invalid file paths, unresolvable operation IDs, missing required slot data, or schema type-validation mismatches. Error details are printed to standard error (`stderr`).

---

## 3. Output Formatting Examples

The `--output` flag allows the engine to serve both human developers debugging in a terminal and automated systems parsing programmatic objects.

### Standard Text Output (Default)
When `--output` is omitted or set to `text`, the CLI prints only the raw, resolved GNAP access array. This is ideal for piping the payload directly into a cURL command or saving it to a file.

**Command:**
```bash
go run . --spec ./fixtures/openapi.yaml --operationId createOutgoingPayment --slots ./fixtures/slots.json
```

**Output:**
```json
[
  {
    "actions": [
      "create"
    ],
    "identifier": "[https://wallet.example.com/alice](https://wallet.example.com/alice)",
    "limits": {
      "debitAmount": {
        "assetCode": "USD",
        "assetScale": 2,
        "value": "5000"
      },
      "receiver": "[https://wallet.example.com/bob/incoming-payments/8b92b6a5-7eb8-4b72-8854-4a460980c986](https://wallet.example.com/bob/incoming-payments/8b92b6a5-7eb8-4b72-8854-4a460980c986)"
    },
    "type": "outgoing-payment"
  }
]
```

### Structured JSON Output
When `--output json` is declared, the CLI wraps the resolved grant in a structured JSON object. This is ideal for SDKs or test runners that need to parse the execution status alongside the payload.

**Command:**
```bash
go run . --spec ./fixtures/openapi.yaml --operationId createOutgoingPayment --slots ./fixtures/slots.json --output json
```

**Output:**
```json
{
  "status": "success",
  "operationId": "createOutgoingPayment",
  "resolvedGrant": [
    {
      "actions": [
        "create"
      ],
      "identifier": "[https://wallet.example.com/alice](https://wallet.example.com/alice)",
      "limits": {
        "debitAmount": {
          "assetCode": "USD",
          "assetScale": 2,
          "value": "5000"
        },
        "receiver": "[https://wallet.example.com/bob/incoming-payments/8b92b6a5-7eb8-4b72-8854-4a460980c986](https://wallet.example.com/bob/incoming-payments/8b92b6a5-7eb8-4b72-8854-4a460980c986)"
      },
      "type": "outgoing-payment"
    }
  ]
}
```
