/*
Copyright 2025 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This tool removes stale translations (keys not in strings.txt) from language JSON files.
// Usage: go run cmd/prune-translations/prune-translations.go (from minikube root)

package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    // Read valid keys from strings.txt
    data, err := os.ReadFile("translations/strings.txt")
    if err != nil {
        panic("Run from minikube root directory. strings.txt not found.")
    }

    var validKeys map[string]interface{}
    if err := json.Unmarshal(data, &validKeys); err != nil {
        panic(err)
    }
    fmt.Printf("Found %d valid keys in strings.txt\n\n", len(validKeys))

    // Process each JSON file
    files, _ := filepath.Glob("translations/*.json")
    for _, path := range files {
        data, err := os.ReadFile(path)
        if err != nil {
            panic(err)
        }

        var translations map[string]interface{}
        if err := json.Unmarshal(data, &translations); err != nil {
            panic(err)
        }

        original := len(translations)
        for key := range translations {
            if _, ok := validKeys[key]; !ok {
                delete(translations, key)
            }
        }

        removed := original - len(translations)
        if removed > 0 {
            output, _ := json.MarshalIndent(translations, "", "\t")
            os.WriteFile(path, append(output, '\n'), 0644)
        }
        fmt.Printf("  %s: %d removed, %d keys (strings.txt: %d)\n",
            filepath.Base(path), removed, len(translations), len(validKeys))
    }
    fmt.Println("\nDone!")
}
