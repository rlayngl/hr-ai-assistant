package main

import "testing"

func TestEthicsAlwaysVetoesOffer(t *testing.T) {
	cases := []Candidate{
		{Experience: true, TestScore: 100, SoftSkills: 100, EthicsPassed: false},
		{Experience: false, TestScore: 100, SoftSkills: 100, EthicsPassed: false},
		{Experience: true, TestScore: 80, SoftSkills: 20, EthicsPassed: false},
	}
	for _, c := range cases {
		if got := EvaluateCandidate(c, DefaultDecisionConfig()); got.Decision != "REJECT" || got.Rule != "R1_ETHICS_VETO" {
			t.Fatalf("ethics veto failed: %+v", got)
		}
	}
}

func TestRulesFromLabOne(t *testing.T) {
	cfg := DefaultDecisionConfig()
	tests := []struct {
		name string
		c    Candidate
		want string
		rule string
	}{
		{"experienced_and_strong", Candidate{Experience: true, TestScore: 90, SoftSkills: 80, EthicsPassed: true}, "OFFER", "R3_EXPERIENCED_STRONG"},
		{"no_experience_but_strong_skills", Candidate{Experience: false, TestScore: 90, SoftSkills: 80, EthicsPassed: true}, "OFFER", "R4_STRONG_SKILLS_NO_EXPERIENCE"},
		{"experienced_low_soft_skills", Candidate{Experience: true, TestScore: 90, SoftSkills: 30, EthicsPassed: true}, "OFFER", "R5_EXPERIENCE_COMPENSATES_SOFT_SKILLS"},
		{"low_test", Candidate{Experience: true, TestScore: 55, SoftSkills: 90, EthicsPassed: true}, "REJECT", "R2_LOW_TECHNICAL_SCORE"},
		{"no_experience_weak_soft_skills", Candidate{Experience: false, TestScore: 90, SoftSkills: 30, EthicsPassed: true}, "REJECT", "R6_NO_EXPERIENCE_WEAK_SOFT_SKILLS"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluateCandidate(tc.c, cfg)
			if got.Decision != tc.want || got.Rule != tc.rule {
				t.Fatalf("got %s/%s; want %s/%s", got.Decision, got.Rule, tc.want, tc.rule)
			}
		})
	}
}

func TestAllDecisionCombinationsEthicsConstraint(t *testing.T) {
	total, _, ethicsOffers := CountDecisionCases()
	if total != 16 {
		t.Fatalf("expected 16 combinations, got %d", total)
	}
	if ethicsOffers != 0 {
		t.Fatalf("ethical veto violated in %d cases", ethicsOffers)
	}
}

func TestResumeClassifier(t *testing.T) {
	got := ClassifyResume("Python Django DRF MySQL RESTful web services Github", false)
	if got.Category != "Python Developer" {
		t.Fatalf("got %q, expected Python Developer", got.Category)
	}
	goGot := ClassifyResume("Go developer golang goroutines channels go modules", false)
	if goGot.Category != "Go Developer" {
		t.Fatalf("got %q, expected Go Developer", goGot.Category)
	}
}
