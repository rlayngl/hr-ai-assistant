package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
)

type ResumeRecord struct{ Category, Resume string }

type ClassMetrics struct {
	Category  string
	Support   int
	Precision float64
	Recall    float64
	F1        float64
}

type Evaluation struct {
	Rows       int
	UniqueRows int
	Correct    int
	Accuracy   float64
	MacroF1    float64
	Classes    []ClassMetrics
	Confusion  map[string]map[string]int
}

func LoadResumeCSV(path string) ([]ResumeRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("CSV must contain header and at least one record")
	}
	header := make([]string, len(rows[0]))
	for i, h := range rows[0] {
		header[i] = strings.ToLower(strings.TrimSpace(h))
	}
	catIdx, textIdx := -1, -1
	for i, h := range header {
		if h == "category" || h == "label" {
			catIdx = i
		}
		if h == "resume" || h == "resume_text" || h == "text" {
			textIdx = i
		}
	}
	if catIdx < 0 || textIdx < 0 {
		return nil, fmt.Errorf("CSV needs Category and Resume columns (case-insensitive)")
	}
	out := make([]ResumeRecord, 0, len(rows)-1)
	for rowNo, row := range rows[1:] {
		if catIdx >= len(row) || textIdx >= len(row) {
			continue
		}
		cat, text := strings.TrimSpace(row[catIdx]), strings.TrimSpace(row[textIdx])
		if cat == "" || text == "" {
			continue
		}
		out = append(out, ResumeRecord{Category: cat, Resume: text})
		_ = rowNo
	}
	return out, nil
}

func EvaluateResumeDataset(records []ResumeRecord, deduplicate bool) Evaluation {
	e := Evaluation{Rows: len(records), Confusion: map[string]map[string]int{}}
	seen := map[string]bool{}
	filtered := make([]ResumeRecord, 0, len(records))
	for _, r := range records {
		key := normalizeText(r.Resume)
		if deduplicate && seen[key] {
			continue
		}
		seen[key] = true
		filtered = append(filtered, r)
	}
	e.UniqueRows = len(filtered)
	trueCount := map[string]int{}
	predCount := map[string]int{}
	tp := map[string]int{}
	classes := map[string]bool{}
	for _, r := range filtered {
		pred := ClassifyResume(r.Resume, false).Category
		classes[r.Category] = true
		classes[pred] = true
		trueCount[r.Category]++
		predCount[pred]++
		if e.Confusion[r.Category] == nil {
			e.Confusion[r.Category] = map[string]int{}
		}
		e.Confusion[r.Category][pred]++
		if pred == r.Category {
			e.Correct++
			tp[r.Category]++
		}
	}
	if len(filtered) > 0 {
		e.Accuracy = float64(e.Correct) / float64(len(filtered))
	}
	keys := make([]string, 0, len(classes))
	for k := range classes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sumF1 := 0.0
	supportClasses := 0
	for _, category := range keys {
		if trueCount[category] == 0 {
			continue
		}
		p, r := 0.0, 0.0
		if predCount[category] > 0 {
			p = float64(tp[category]) / float64(predCount[category])
		}
		if trueCount[category] > 0 {
			r = float64(tp[category]) / float64(trueCount[category])
		}
		f1 := 0.0
		if p+r > 0 {
			f1 = 2 * p * r / (p + r)
		}
		e.Classes = append(e.Classes, ClassMetrics{Category: category, Support: trueCount[category], Precision: p, Recall: r, F1: f1})
		sumF1 += f1
		supportClasses++
	}
	if supportClasses > 0 {
		e.MacroF1 = sumF1 / float64(supportClasses)
	}
	return e
}

type SweepRow struct {
	TechThreshold, SoftThreshold float64
	Candidates, Offers           int
	OfferRate                    float64
}

func RunSensitivitySweep() []SweepRow {
	techThresholds := []float64{50, 60, 70, 80, 90}
	softThresholds := []float64{40, 50, 60, 70, 80}
	rows := []SweepRow{}
	for _, tt := range techThresholds {
		for _, st := range softThresholds {
			offers, n := 0, 0
			for exp := 0; exp <= 1; exp++ {
				for ethics := 0; ethics <= 1; ethics++ {
					for test := 40.0; test <= 100; test += 10 {
						for soft := 30.0; soft <= 100; soft += 10 {
							c := Candidate{Experience: exp == 1, TestScore: test, SoftSkills: soft, EthicsPassed: ethics == 1}
							d := EvaluateCandidate(c, DecisionConfig{TechnicalThreshold: tt, SoftSkillsThreshold: st})
							if d.Decision == "OFFER" {
								offers++
							}
							n++
						}
					}
				}
			}
			rows = append(rows, SweepRow{TechThreshold: tt, SoftThreshold: st, Candidates: n, Offers: offers, OfferRate: float64(offers) / float64(n)})
		}
	}
	return rows
}

func WriteSweepCSV(path string, rows []SweepRow) error {
	if err := os.MkdirAll(parentDir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	_ = w.Write([]string{"technical_threshold", "soft_skills_threshold", "synthetic_cases", "offers", "offer_rate"})
	for _, r := range rows {
		_ = w.Write([]string{fmt.Sprint(r.TechThreshold), fmt.Sprint(r.SoftThreshold), fmt.Sprint(r.Candidates), fmt.Sprint(r.Offers), fmt.Sprintf("%.6f", r.OfferRate)})
	}
	return w.Error()
}

func parentDir(path string) string {
	i := strings.LastIndexAny(path, `/\\`)
	if i < 0 {
		return "."
	}
	if i == 0 {
		return string(path[0])
	}
	return path[:i]
}

func minMaxSweep(rows []SweepRow) (min, max SweepRow) {
	if len(rows) == 0 {
		return
	}
	min, max = rows[0], rows[0]
	for _, r := range rows[1:] {
		if r.OfferRate < min.OfferRate {
			min = r
		}
		if r.OfferRate > max.OfferRate {
			max = r
		}
	}
	return
}

func rounded(x float64) float64 { return math.Round(x*10000) / 10000 }

func CountDecisionCases() (total, offers, ethicsFailOffers int) {
	cfg := DefaultDecisionConfig()
	for exp := 0; exp < 2; exp++ {
		for soft := 0; soft < 2; soft++ {
			for test := 0; test < 2; test++ {
				for ethics := 0; ethics < 2; ethics++ {
					c := Candidate{Experience: exp == 1, SoftSkills: map[bool]float64{true: 80, false: 30}[soft == 1], TestScore: map[bool]float64{true: 90, false: 50}[test == 1], EthicsPassed: ethics == 1}
					d := EvaluateCandidate(c, cfg)
					total++
					if d.Decision == "OFFER" {
						offers++
						if !c.EthicsPassed {
							ethicsFailOffers++
						}
					}
				}
			}
		}
	}
	return
}

func writeString(w io.Writer, s string) { _, _ = io.WriteString(w, s) }
