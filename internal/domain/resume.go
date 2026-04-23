package domain

// StructuredResume 解析后存 Redis key: resume:{user_id}
type StructuredResume struct {
	UserID      string             `json:"user_id"`
	Skills      []string           `json:"skills"`
	Projects    []ResumeProject    `json:"projects"`
	Internships []ResumeInternship `json:"internships"`
	Education   ResumeEducation    `json:"education"`
}

type ResumeProject struct {
	Name        string   `json:"name"`
	TechStack   []string `json:"tech_stack"`
	Description string   `json:"description"`
	Highlights  []string `json:"highlights"`
}

type ResumeInternship struct {
	Company     string `json:"company,omitempty"`
	Role        string `json:"role,omitempty"`
	Description string `json:"description,omitempty"`
}

type ResumeEducation struct {
	School     string `json:"school"`
	Major      string `json:"major"`
	Graduation string `json:"graduation"`
}
