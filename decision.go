package main

import "fmt"

type Candidate struct {
	ID           string  `json:"id,omitempty"`
	Experience   bool    `json:"commercial_experience"`
	TestScore    float64 `json:"technical_test_score"`
	SoftSkills   float64 `json:"soft_skills_score"`
	EthicsPassed bool    `json:"ethics_and_security_passed"`
	ResumeText   string  `json:"resume_text,omitempty"`
}

type DecisionConfig struct {
	TechnicalThreshold  float64 `json:"technical_threshold"`
	SoftSkillsThreshold float64 `json:"soft_skills_threshold"`
}

func DefaultDecisionConfig() DecisionConfig {
	return DecisionConfig{TechnicalThreshold: 70, SoftSkillsThreshold: 60}
}

type DecisionResult struct {
	Decision string   `json:"decision"`
	Rule     string   `json:"rule"`
	Reasons  []string `json:"reasons"`
	Role     string   `json:"predicted_resume_category,omitempty"`
}

// EvaluateCandidate is a deterministic, interpretable production-rule model
// based on the decision rules from Laboratory Work 1.
func EvaluateCandidate(c Candidate, cfg DecisionConfig) DecisionResult {
	if cfg.TechnicalThreshold < 0 || cfg.TechnicalThreshold > 100 {
		cfg.TechnicalThreshold = 70
	}
	if cfg.SoftSkillsThreshold < 0 || cfg.SoftSkillsThreshold > 100 {
		cfg.SoftSkillsThreshold = 60
	}
	if !c.EthicsPassed {
		return DecisionResult{Decision: "REJECT", Rule: "R1_ETHICS_VETO", Reasons: []string{"Критерий этики/безопасности не пройден: срабатывает безусловный запрет."}}
	}
	if c.TestScore < cfg.TechnicalThreshold {
		return DecisionResult{Decision: "REJECT", Rule: "R2_LOW_TECHNICAL_SCORE", Reasons: []string{fmt.Sprintf("Технический балл %.1f ниже порога %.1f.", c.TestScore, cfg.TechnicalThreshold)}}
	}
	if c.SoftSkills >= cfg.SoftSkillsThreshold {
		reason := fmt.Sprintf("Технический балл %.1f не ниже порога %.1f; soft skills %.1f не ниже порога %.1f; этический критерий пройден.", c.TestScore, cfg.TechnicalThreshold, c.SoftSkills, cfg.SoftSkillsThreshold)
		if c.Experience {
			return DecisionResult{Decision: "OFFER", Rule: "R3_EXPERIENCED_STRONG", Reasons: []string{reason, "Коммерческий опыт подтверждён."}}
		}
		return DecisionResult{Decision: "OFFER", Rule: "R4_STRONG_SKILLS_NO_EXPERIENCE", Reasons: []string{reason, "Опыт не является абсолютным условием: сильные технические и коммуникативные результаты позволяют дать шанс кандидату без опыта."}}
	}
	if c.Experience {
		return DecisionResult{Decision: "OFFER", Rule: "R5_EXPERIENCE_COMPENSATES_SOFT_SKILLS", Reasons: []string{fmt.Sprintf("Технический балл %.1f не ниже порога %.1f, коммерческий опыт подтверждён, этика пройдена.", c.TestScore, cfg.TechnicalThreshold), fmt.Sprintf("Soft skills %.1f ниже порога %.1f, но по принятому правилу опытный кандидат с сильным тестом может быть принят на техническую позицию.", c.SoftSkills, cfg.SoftSkillsThreshold)}}
	}
	return DecisionResult{Decision: "REJECT", Rule: "R6_NO_EXPERIENCE_WEAK_SOFT_SKILLS", Reasons: []string{fmt.Sprintf("Технический балл %.1f достаточен, но коммерческого опыта нет и soft skills %.1f ниже порога %.1f.", c.TestScore, c.SoftSkills, cfg.SoftSkillsThreshold), "Сочетание факторов соответствует правилу отказа для кандидата без опыта."}}
}
