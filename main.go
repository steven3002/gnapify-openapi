package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	specPath := flag.String("spec", "", "Path to the openapi.yaml (Required)")
	operationID := flag.String("operationId", "", "The operationId (Required)")
	slotsPath := flag.String("slots", "", "Path to slots.json (Optional)")
	outputFormat := flag.String("output", "text", "Output format: 'text' or 'json'")

	flag.Parse()

	if *specPath == "" || *operationID == "" {
		fmt.Fprintln(os.Stderr, "Error: --spec and --operationId are required.")
		os.Exit(1)
	}

	doc, err := LoadSpec(*specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %v\n", err)
		os.Exit(1)
	}

	securityArray, err := FindOperationSecurity(doc, *operationID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: operationId '%s' not found in specification or has no security.\n", *operationID)
		os.Exit(1) // Matches TC-3.4
	}

	slotsData, err := LoadSlots(*slotsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %v\n", err)
		os.Exit(1)
	}

	// Iterate through the security array to handle OR logic gracefully (TC-3.5)
	var lastError error
	for _, secReqRaw := range securityArray {
		secReq, ok := toMap(secReqRaw)
		if !ok {
			continue
		}

		gnapReqsRaw, hasGNAP := secReq["GNAP"]
		if !hasGNAP {
			continue
		}

		gnapReqs, ok := gnapReqsRaw.([]interface{})
		if !ok || len(gnapReqs) == 0 {
			continue
		}

		profileName := fmt.Sprintf("%v", gnapReqs[0])

		// Fetch the profile blueprint
		profileTemplate, err := FindGNAPProfile(doc, profileName)
		if err != nil {
			lastError = err
			continue
		}

		// Resolve the blueprint with slots AND the OpenAPI doc for type checking
		resolvedProfile, err := ResolveNode(profileTemplate, slotsData, doc)
		if err != nil {
			lastError = err
			continue // Try the next security requirement if this one fails
		}

		// SUCCESS! Format and print the result
		outputJSON, _ := json.MarshalIndent(resolvedProfile, "", "  ")

		if *outputFormat == "json" {
			fmt.Printf("{\n  \"status\": \"success\",\n  \"operationId\": \"%s\",\n  \"resolvedGrant\": %s\n}\n", *operationID, string(outputJSON))
		} else {
			fmt.Println(string(outputJSON))
		}
		os.Exit(0)
	}

	// If it exhausted all options without success, print the last error
	fmt.Fprintf(os.Stderr, "Resolution Error: %v\n", lastError)
	os.Exit(1)
}
