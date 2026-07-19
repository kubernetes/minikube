package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/minikube/cmd/minikube/cmd/flags"
	doctor "k8s.io/minikube/pkg/minikube/doctor"
	"k8s.io/minikube/pkg/minikube/mustload"
)

var doctorVerbose bool
var doctorOutput string

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostics on your Minikube installation",
	Long:  "Run diagnostics on your Minikube installation.",

	Run: func(cmd *cobra.Command, args []string) {
		options := flags.CommandOptions()
		cname := ClusterFlagValue()
		api, cc := mustload.Partial(cname, options)

		results := doctor.Run(api, cc)

		if doctorOutput == "json" {
			data, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				fmt.Printf("Error formatting JSON: %v\n", err)
				return
			}
			fmt.Println(string(data))
			return
		}

		fmt.Println()
		fmt.Println("Minikube Doctor")
		fmt.Println("────────────────────────────────────")

		fmt.Println()
		fmt.Println("Configuration")
		if len(results) >= 4 {
			fmt.Printf("  %-20s: %s\n", results[0].Name, results[0].Message)
			fmt.Printf("  %-20s: %s\n", results[1].Name, results[1].Message)
			fmt.Printf("  %-20s: %s\n", results[2].Name, results[2].Message)
			fmt.Printf("  %-20s: %s\n", results[3].Name, results[3].Message)
		}

		printCheck := func(r doctor.Result) {
			symbol := "✓"
			if r.Status == "FAIL" {
				symbol = "✗"
			} else if r.Status == "WARNING" {
				symbol = "⚠"
			}
			fmt.Printf("%s %-20s : %s (%s)\n", symbol, r.Name, r.Message, r.Status)
			if r.Status == "FAIL" || r.Status == "WARNING" || doctorVerbose {
				if r.Details != "" && doctorVerbose {
					fmt.Printf("    Details: %s\n", r.Details)
				}
				if r.Recommendation != "" {
					fmt.Printf("    Recommendation: %s\n", r.Recommendation)
				}
			}
		}

		if len(results) > 11 {
			fmt.Println()
			fmt.Println("Validation")
			printCheck(results[11])
		}

		fmt.Println()
		fmt.Println("Environment")
		for i := 4; i < 7 && i < len(results); i++ {
			printCheck(results[i])
		}

		fmt.Println()
		fmt.Println("Cluster")
		for i := 7; i < 11 && i < len(results); i++ {
			printCheck(results[i])
		}

		fmt.Println()
		fmt.Println("Resources")
		for i := 12; i < len(results); i++ {
			printCheck(results[i])
		}

		fmt.Println()
		fmt.Println("Summary")
		fmt.Println("────────────────────────────────────")
		passed := 0
		warnings := 0
		failed := 0
		// Exclude first 4 informational fields from test counts
		checkResults := results[4:]
		if len(results) > 11 {
			checkResults = append(results[4:11], results[11:]...)
		}
		for _, r := range checkResults {
			if r.Status == "PASS" {
				passed++
			} else if r.Status == "WARNING" {
				warnings++
			} else {
				failed++
			}
		}

		totalChecks := len(checkResults)
		healthScore := 100
		if totalChecks > 0 {
			// Count warnings as half-passed
			healthScore = int((float64(passed) + float64(warnings)*0.5) / float64(totalChecks) * 100)
		}

		healthStatus := "Healthy"
		if healthScore < 50 {
			healthStatus = "Unhealthy"
		} else if healthScore < 90 {
			healthStatus = "Needs Attention"
		}

		fmt.Printf("Overall Health  : %d%% (%s)\n", healthScore, healthStatus)
		fmt.Printf("Checks Run      : %d\n", totalChecks)
		fmt.Printf("Passed          : %d\n", passed)
		fmt.Printf("Warnings        : %d\n", warnings)
		fmt.Printf("Failed          : %d\n", failed)

		if doctorVerbose {
			fmt.Println()
			fmt.Println("Diagnostics completed.")
		}
	},
}

func init() {
	doctorCmd.Flags().BoolVar(
		&doctorVerbose,
		"verbose",
		false,
		"Display detailed diagnostic information",
	)
	doctorCmd.Flags().StringVarP(
		&doctorOutput,
		"output",
		"o",
		"text",
		"Format output: 'text' or 'json'",
	)
}
