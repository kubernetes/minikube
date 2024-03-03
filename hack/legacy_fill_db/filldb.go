/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres" // Blank import used for registering cloudsql driver as a database driver
	"github.com/jmoiron/sqlx"
)

const (
	bucketPath    = "gs://minikube-builds/logs/master"
	dbBackend     = "postgres"
	mkRepo        = "github.com/kubernetes/minikube/"
	host          = "k8s-minikube:us-west1:flake-rate"
	dbPathPattern = "user=postgres dbname=flaketest2 password=%s"
	gopoghCommand = "%s -name '%s' -repo '%s' -pr 'HEAD' -in '%s' -out_html './out/output.html' -out_summary out/output_summary.json -details '%s' -use_cloudsql -db_host '%s' -db_path '%s' -db_backend '%s'"
)

var rateLimitExceeded = false
var existingEnvironments = make(map[string]map[string]struct{})

func main() {
	gp, err := exec.LookPath("gopogh")
	if err != nil {
		log.Fatalf("missing gopogh. Run 'go install github.com/medyagh/gopogh/cmd/gopogh@latest': %v", err)
	}
	f, err := os.OpenFile("gopogh_filldb_log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open/create gopogh_filldb_log: %v", err)
	}
	logger := log.New(f, "", log.LstdFlags)
	var wg sync.WaitGroup
	const numPartitions = 64

	populateExistingEnvironments()
	for i := 0; i < numPartitions; i++ {
		wg.Add(1)
		go func(partition int) {
			defer wg.Done()

			// Calculate the partition range
			folders, err := listFolders(partition, numPartitions)
			if err != nil {
				fmt.Printf("Error listing folders: %v", err)
				return
			}

			// Iterate over the commit folders in this partition
			for _, folder := range folders {
				// Extract the commit SHA (or whatever it is?) from the folder path
				commitSha := filepath.Base(folder)

				// Process the commit folder
				err := processCommitFolder(commitSha, logger, gp)
				if err != nil {
					fmt.Printf("Error processing commit folder %s: %v", commitSha, err)
					if !strings.Contains(err.Error(), "failed to copy jsons. may not exist for folder") {
						logger.Printf("Failed to process commit folder: %s. Error: %v\n\n", commitSha, err)
					}
				}
			}
		}(i)
	}

	wg.Wait()
}

func listFolders(partition, numPartitions int) ([]string, error) {
	// Use the 'gsutil' command to list the folders and then partition them
	cmd := exec.Command("gsutil", "ls", "-lah", bucketPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var folders []string
	for i, line := range lines {
		if i%(numPartitions) == partition {
			folders = append(folders, strings.TrimSpace(line))
		}
	}

	return folders, nil
}

func processCommitFolder(commitSha string, logger *log.Logger, gp string) error {
	// Create a local directory to store the JSON files
	dirPath := fmt.Sprintf("/tmp/gopogh-fill-db/%s", commitSha)
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dirPath)

	// Copy all JSON files from the folder to the local directory
	copyCmd := exec.Command("gsutil", "cp", fmt.Sprintf("%s/%s/*.json", bucketPath, commitSha), dirPath)
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy jsons. may not exist for folder. %v", err)
	}

	// Get the summary JSON filename
	summaryFilename, err := findSummaryJSON(dirPath)
	if err != nil {
		return fmt.Errorf("failed to get summary json: %v", err)
	}

	// If a summary JSON file exists, extract the commitSha from the "Details" field
	if summaryFilename != "" {
		jsonData, err := os.ReadFile(summaryFilename)
		if err != nil {
			return err
		}

		var summaryData map[string]interface{}
		if err := json.Unmarshal(jsonData, &summaryData); err != nil {
			return err
		}

		if details, ok := summaryData["Detail"].(map[string]interface{}); ok {
			if commitShaVal, ok := details["Details"].(string); ok {
				commitSha = commitShaVal
			}
		}

		if dur, ok := summaryData["TotalDuration"].(string); ok && dur == "0" {
			return err
		}
	}

	dbPath := fmt.Sprintf(dbPathPattern, os.Getenv("DB_PASS"))
	// Iterate over the JSON files in the local directory
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, _ error) error {
		filename := filepath.Base(path[:len(path)-len(".json")])
		if !info.IsDir() && strings.HasSuffix(path, ".json") && !strings.HasSuffix(path, "summary.json") && notInDB(commitSha, filename) {
			gopoghCmd := fmt.Sprintf(gopoghCommand, gp, filename, mkRepo, path, commitSha, host, dbPath, dbBackend)

			for rateLimitExceeded {
				time.Sleep(time.Hour)
			}
			err := runAndRetry(gopoghCmd, logger)
			if err != nil {
				fmt.Printf("Error executing gopogh command for %s: %v\n", filename, err)
				logger.Printf("Failed to run gopogh on %s \n", filename)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func notInDB(commit string, env string) bool {
	if _, ok := existingEnvironments[env][commit]; ok {
		fmt.Printf("Environment '%s' with commit SHA '%s' already exists in the database. Skipping...\n", env, commit)
		return false
	}
	return true
}

func runAndRetry(gopoghCmd string, logger *log.Logger) error {
	retries := 100
	retryInterval := time.Hour

	for i := 0; i < retries; i++ {
		gopogh := exec.Command("sh", "-c", gopoghCmd)
		if o, err := gopogh.CombinedOutput(); err == nil && !strings.Contains(string(o), "RATE_LIMIT_EXCEEDED") {
			rateLimitExceeded = false
			return nil // Execution succeeded
		} else if strings.Contains(string(o), "RATE_LIMIT_EXCEEDED") {
			fmt.Printf("Rate limit exceeded. Retrying after 1 hour. Time now is: %s \n", time.Now().String())
			rateLimitExceeded = true
			time.Sleep(retryInterval)
		} else if strings.Contains(string(o), "connect: connection timed out") {
			fmt.Printf("Connection timed out")
			time.Sleep(5 * time.Minute)
		} else {
			logger.Printf("Error running gopogh: %s\n", string(o))
			logger.Printf("Command was: %s\n\n", gopoghCmd)
			rateLimitExceeded = false
			return err // Other error occurred, return immediately
		}
	}
	logger.Printf("Error running gopogh. Exceeded maximum number of retries\n")
	logger.Printf("Command was: %s\n\n", gopoghCmd)
	return fmt.Errorf("exceeded maximum number of retries")
}

func findSummaryJSON(dirPath string) (string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), "summary.json") {
			return filepath.Join(dirPath, file.Name()), nil
		}
	}

	return "", nil
}

func populateExistingEnvironments() {
	dbPath := fmt.Sprintf(dbPathPattern, os.Getenv("DB_PASS"))
	connectString := fmt.Sprintf("host=%s %s sslmode=disable", host, dbPath)

	dbx, err := sqlx.Connect("cloudsqlpostgres", connectString)
	if err != nil {
		log.Fatalf("connect failed: %v", err)
	}
	rows, err := dbx.Query("SELECT envname, commitid FROM db_environment_tests")
	if err != nil {
		log.Fatalf("failed to select environment table: %v", err)
	}
	for rows.Next() {
		var environmentName, commitSHA string
		if err := rows.Scan(&environmentName, &commitSHA); err != nil {
			closeErr := rows.Close()
			if closeErr != nil {
				log.Printf("failed to close rows: %v", closeErr)
			}
			log.Fatalf("failed to parse env table: %v", err)
		}
		if _, ok := existingEnvironments[environmentName]; !ok {
			existingEnvironments[environmentName] = make(map[string]struct{})
		}
		existingEnvironments[environmentName][commitSHA] = struct{}{}
	}
	if closeErr := rows.Close(); closeErr != nil {
		log.Printf("failed to close rows: %v", closeErr)
	}
	if err = dbx.Close(); err != nil {
		log.Fatalf("failed to close database connection: %v", err)
	}
}
