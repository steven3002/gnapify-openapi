# Writing Tests: The Sandbox Guide

This  provides a step-by-step guide for developers to write custom tests, experiment with GNAP profiles, and manipulate the Conformance Engine. It explains how to add new fixtures, inject dynamic data, and perform regression testing.

## 1. Adding a New OpenAPI Fixture

To test a new GNAP profile, an OpenAPI document (fixture) must be created. All test documents reside in the `fixtures/` directory. 

Create a new file named `openapi-custom-test.yaml` inside the `fixtures/` directory. Define a path, an operation ID, and a GNAP security requirement. Ensure the requested profile exists in the `x-gnap-access-profiles` catalog.

**Example `fixtures/openapi-custom-test.yaml`:**
```yaml
openapi: 3.1.0
info:
  title: Custom GNAP Test
  version: 1.0.0
paths:
  /custom-endpoint:
    post:
      operationId: customTestOperation
      responses:
        '200':
          description: Success
      security:
        - GNAP: ["customDynamicProfile"]
components:
  securitySchemes:
    GNAP:
      type: http
      scheme: GNAP
      x-gnap-access-profiles:
        customDynamicProfile:
          - type: "custom-access"
            actions: ["read", "write"]
            accountId:
              $slot: "user.accountId"
              schemaRef: "#/components/schemas/String"
```

## 2. Mocking Dynamic Data (slots.json)

If the newly created profile is dynamic (contains `$slot` properties), a JSON file containing the mock data must be provided. 

Create a file named `custom-slots.json` inside the `fixtures/` directory. The keys in this JSON file must exactly match the `$slot` strings defined in the OpenAPI fixture.

**Example `fixtures/custom-slots.json`:**
```json
{
  "user.accountId": "acc_123456789"
}
```

## 3. Verifying Against the Validation Layer

Before running the resolver engine, the new OpenAPI fixture must pass the static contract rules. Use the Spectral linter to verify structural integrity and reference accuracy.

Run the following command:
```bash
spectral lint fixtures/openapi-custom-test.yaml
```

A successful test will output zero errors. If typographical errors exist in the `$slot` definitions or if the profile reference is missing, the linter will immediately flag them.

## 4. Executing the Conformance Engine

Once the fixture passes validation, the Go CLI can be used to test the resolution logic. The engine will merge the OpenAPI template with the mock JSON data.

Run the following command:
```bash
go run . --spec ./fixtures/openapi-custom-test.yaml --operationId customTestOperation --slots ./fixtures/custom-slots.json
```

The engine should output the resolved GNAP access array. Verify that the dynamic data injected cleanly into the correct fields. 

**Expected Output:**
```json
[
  {
    "actions": [
      "read",
      "write"
    ],
    "accountId": "acc_123456789",
    "type": "custom-access"
  }
]
```

## 5. Regression Testing and the Corpus

The `corpus/` directory acts as the strict baseline for the project. It contains the "known good" resolved JSON outputs for the core Open Payments profiles.

When modifying the Go resolution engine (`resolver.go`) or the Spectral rules, the output of the CLI must be checked against the files in the `corpus/` directory to ensure no breaking changes were introduced.

To perform a regression test:

1. Route the CLI output of a core fixture to a temporary file.
2. Compare the temporary file against the corresponding file in the `corpus/` directory using a standard `diff` command.

**Example Regression Command:**
```bash
go run . --spec ./fixtures/openapi.yaml --operationId createOutgoingPayment --slots ./fixtures/slots.json > temp.json
diff temp.json corpus/outgoingPaymentCreateFixedDebit.json
```

If the `diff` command returns no output, the engine changes are safe and the contract remains intact.
