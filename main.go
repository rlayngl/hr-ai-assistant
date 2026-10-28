package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	mode := flag.String("mode", "demo", "mode: demo | classify | decision | eval | sweep")
	text := flag.String("text", "", "resume text for classification")
	input := flag.String("input", "", "candidate JSON file for decision mode")
	dataset := flag.String("dataset", "testdata/resume_benchmark.csv", "CSV with Category,Resume columns")
	tech := flag.Float64("tech", 70, "technical threshold 0..100")
	soft := flag.Float64("soft", 60, "soft-skills threshold 0..100")
	out := flag.String("out", "results/parameter_sweep.csv", "sensitivity-sweep CSV output")
	flag.Parse()

	switch strings.ToLower(*mode) {
	case "classify":
		if strings.TrimSpace(*text) == "" {
			fatal("-text is required in classify mode")
		}
		printJSON(ClassifyResume(*text, true))
	case "decision":
		if *input == "" {
			fatal("-input is required in decision mode")
		}
		var c Candidate
		b, err := os.ReadFile(*input)
		if err != nil {
			fatal(err.Error())
		}
		if err := json.Unmarshal(b, &c); err != nil {
			fatal("invalid candidate JSON: " + err.Error())
		}
		cfg := DecisionConfig{TechnicalThreshold: *tech, SoftSkillsThreshold: *soft}
		if cfg.TechnicalThreshold == 0 {
			cfg.TechnicalThreshold = 70
		}
		if cfg.SoftSkillsThreshold == 0 {
			cfg.SoftSkillsThreshold = 60
		}
		result := EvaluateCandidate(c, cfg)
		if strings.TrimSpace(c.ResumeText) != "" {
			result.Role = ClassifyResume(c.ResumeText, false).Category
		}
		printJSON(result)
	case "eval":
		records, err := LoadResumeCSV(*dataset)
		if err != nil {
			fatal(err.Error())
		}
		printEvaluation(*dataset, EvaluateResumeDataset(records, true))
	case "sweep":
		rows := RunSensitivitySweep()
		if err := WriteSweepCSV(*out, rows); err != nil {
			fatal(err.Error())
		}
		mn, mx := minMaxSweep(rows)
		fmt.Printf("Sensitivity sweep completed: %d configurations; %d candidate profiles per configuration.\n", len(rows), mn.Candidates)
		fmt.Printf("Lowest offer rate: %.2f%% (technical threshold %.0f, soft-skills threshold %.0f).\n", mn.OfferRate*100, mn.TechThreshold, mn.SoftThreshold)
		fmt.Printf("Highest offer rate: %.2f%% (technical threshold %.0f, soft-skills threshold %.0f).\n", mx.OfferRate*100, mx.TechThreshold, mx.SoftThreshold)
		fmt.Printf("CSV saved to %s\n", *out)
	case "demo":
		fmt.Println("=== Resume classification demo ===")
		printJSON(ClassifyResume("Python, Django, DRF, MySQL, RESTful web services, GitHub, Unix", true))
		fmt.Println("\n=== Candidate decision demo ===")
		c := Candidate{ID: "demo-001", Experience: false, TestScore: 88, SoftSkills: 74, EthicsPassed: true, ResumeText: "Go developer; golang; goroutines; channels; go modules; PostgreSQL; Docker; Kubernetes; REST API"}
		decision := EvaluateCandidate(c, DefaultDecisionConfig())
		decision.Role = ClassifyResume(c.ResumeText, false).Category
		printJSON(decision)
		fmt.Println("\n=== External excerpt benchmark ===")
		records, err := LoadResumeCSV(*dataset)
		if err != nil {
			fatal(err.Error())
		}
		printEvaluation(*dataset, EvaluateResumeDataset(records, true))
		rows := RunSensitivitySweep()
		if err := WriteSweepCSV(*out, rows); err != nil {
			fatal(err.Error())
		}
		mn, mx := minMaxSweep(rows)
		fmt.Printf("\nSensitivity sweep: %d configurations, %d profiles/configuration; offer rate range %.2f%%..%.2f%%.\n", len(rows), mn.Candidates, mn.OfferRate*100, mx.OfferRate*100)
		fmt.Printf("Sweep CSV: %s\n", *out)
	default:
		fatal("unknown mode. Use demo, classify, decision, eval or sweep")
	}
}

func printJSON(v any) {
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	if err := e.Encode(v); err != nil {
		fatal(err.Error())
	}
}
func printEvaluation(path string, e Evaluation) {
	fmt.Printf("Dataset: %s\n", path)
	fmt.Printf("Rows: %d; after duplicate-text removal: %d\n", e.Rows, e.UniqueRows)
	fmt.Printf("Correct: %d; accuracy: %.2f%%; macro-F1: %.4f\n", e.Correct, e.Accuracy*100, e.MacroF1)
	fmt.Println("Per-class metrics:")
	for _, c := range e.Classes {
		fmt.Printf("  %-26s n=%-3d precision=%.3f recall=%.3f F1=%.3f\n", c.Category, c.Support, c.Precision, c.Recall, c.F1)
	}
}
func fatal(s string) { fmt.Fprintln(os.Stderr, "error:", s); os.Exit(2) }
