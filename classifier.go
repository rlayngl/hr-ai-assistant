package main

import (
	"regexp"
	"sort"
	"strings"
)

// CategoryRule is a transparent rule for routing a resume to a job category.
// A score is the number of distinct category keywords found in the resume text.
type CategoryRule struct {
	Category string
	Keywords []string
}

type CategoryScore struct {
	Category string   `json:"category"`
	Score    int      `json:"score"`
	Matched  []string `json:"matched_keywords"`
}

type Classification struct {
	Category   string          `json:"category"`
	Score      int             `json:"score"`
	Confidence float64         `json:"score_share"`
	RunnerUp   string          `json:"runner_up,omitempty"`
	Scores     []CategoryScore `json:"scores,omitempty"`
	Note       string          `json:"note"`
}

var separatorRE = regexp.MustCompile(`[^a-z0-9+#]+`)
var spacesRE = regexp.MustCompile(`\s+`)

func normalizeText(s string) string {
	s = strings.ToLower(s)
	s = separatorRE.ReplaceAllString(s, " ")
	return " " + strings.TrimSpace(spacesRE.ReplaceAllString(s, " ")) + " "
}

func containsTerm(normalizedText, keyword string) bool {
	k := normalizeText(keyword)
	return strings.Contains(normalizedText, k)
}

// CategoryRules covers job domains from the public UpdatedResumeDataSet.
// Keywords are intentionally inspectable and should be tuned with HR/domain experts.
var CategoryRules = []CategoryRule{
	{"Advocate", []string{"advocate", "legal drafting", "plaint", "litigation", "legal cases", "law degree", "income tax returns", "gst returns"}},
	{"Arts", []string{"arts", "sociology", "fine arts", "creative arts", "arts course", "art direction"}},
	{"Automation Testing", []string{"automation testing", "selenium", "webdriver", "testng", "cypress", "page object model", "test automation", "jenkins"}},
	{"Blockchain", []string{"blockchain", "solidity", "smart contract", "ethereum", "web3", "hyperledger"}},
	{"Business Analyst", []string{"business analyst", "requirements gathering", "stakeholder", "user stories", "brd", "process mapping", "mis reports", "uat", "business process"}},
	{"Civil Engineer", []string{"civil engineering", "civil engineer", "autocad", "construction", "structural design", "site engineer", "qa qc civil"}},
	{"Data Science", []string{"data science", "machine learning", "scikit learn", "scipy", "pandas", "numpy", "statistics", "tableau", "data visualization", "deep learning", "nlp", "predictive analysis"}},
	{"Database", []string{"database administrator", "database servers", "sql server", "oracle database", "database management", "t sql", "mysql", "postgresql"}},
	{"DevOps Engineer", []string{"devops", "ci cd", "jenkins", "ansible", "docker", "kubernetes", "maven", "shell scripting", "configuration management", "gitlab ci"}},
	{"DotNet Developer", []string{"dotnet", "c#", "asp net", "aspnet", "entity framework", "visual studio", "sql server", "mvc", "telerik"}},
	{"ETL Developer", []string{"etl developer", "talend", "informatica powercenter", "data integration", "data warehouse", "redshift", "athena", "s3 schema", "star schema"}},
	{"Electrical Engineering", []string{"electrical engineering", "electrical engineer", "power systems", "electrical circuits", "plc", "matlab", "electrical machinery", "fire fighting"}},
	{"HR", []string{"human resources", "hr operations", "recruitment", "talent acquisition", "employee relations", "payroll", "onboarding", "hr manager", "personnel management"}},
	{"Hadoop", []string{"hadoop", "hdfs", "mapreduce", "sqoop", "hive", "hbase", "spark", "big data"}},
	{"Health and fitness", []string{"health and fitness", "nutrition", "exercise physiology", "fitness instructor", "dietitian", "personal trainer", "wellness"}},
	{"Go Developer", []string{"go developer", "golang", "goroutines", "channels", "go modules", "gofmt", "grpc", "gin framework", "fiber framework"}},
	{"Java Developer", []string{"java developer", "spring mvc", "hibernate", "jdbc", "j2ee", "servlet", "struts", "maven", "rest api", "tomcat"}},
	{"Mechanical Engineer", []string{"mechanical engineering", "mechanical engineer", "mechanical design", "cad design", "solidworks", "thermodynamics", "manufacturing process"}},
	{"Network Security Engineer", []string{"network security", "firewall", "penetration testing", "soc analyst", "intrusion detection", "vulnerability assessment", "cyber security", "network security engineer"}},
	{"Operations Manager", []string{"operations manager", "enterprise resource planning", "erp", "employee resource group", "operations management", "tally", "resource planning", "operational efficiency"}},
	{"PMO", []string{"pmo", "project management office", "project status", "project governance", "project portfolio", "senior executive pmo", "project schedule"}},
	{"Python Developer", []string{"python developer", "django", "fastapi", "flask", "drf", "restful web services", "python programming", "pytest", "pip"}},
	{"SAP Developer", []string{"sap developer", "sap consultant", "sap hana", "sap abap", "sap erp", "sap bw", "sap modules"}},
	{"Sales", []string{"sales manager", "sales executive", "business development", "sales target", "lead generation", "customer acquisition", "sales operations", "sales and marketing", "customer sales management", "direct sales management", "sales skills", "convincing power"}},
	{"Testing", []string{"software testing", "manual testing", "test cases", "defect tracking", "functional testing", "regression testing", "quality assurance", "test execution"}},
	{"Web Designing", []string{"web designing", "web design", "html", "css", "responsive design", "website", "bootstrap", "figma", "user interface", "ui ux"}},
}

func ClassifyResume(text string, includeScores bool) Classification {
	n := normalizeText(text)
	scores := make([]CategoryScore, 0, len(CategoryRules))
	for _, rule := range CategoryRules {
		cs := CategoryScore{Category: rule.Category}
		for _, keyword := range rule.Keywords {
			if containsTerm(n, keyword) {
				cs.Score++
				cs.Matched = append(cs.Matched, keyword)
			}
		}
		scores = append(scores, cs)
	}
	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Category < scores[j].Category
		}
		return scores[i].Score > scores[j].Score
	})
	result := Classification{Category: "Unclassified", Note: "Rule-based routing by matching keywords; score is not a probability."}
	if len(scores) > 0 && scores[0].Score > 0 {
		result.Category = scores[0].Category
		result.Score = scores[0].Score
		total := 0
		for _, s := range scores {
			total += s.Score
		}
		if total > 0 {
			result.Confidence = float64(scores[0].Score) / float64(total)
		}
		if len(scores) > 1 && scores[1].Score > 0 {
			result.RunnerUp = scores[1].Category
		}
	}
	if includeScores {
		// Keep only non-zero scores to make CLI output useful.
		for _, s := range scores {
			if s.Score > 0 {
				result.Scores = append(result.Scores, s)
			}
		}
	}
	return result
}
