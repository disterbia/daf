package core

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Jwt string `json:"jwt,omitempty"`
	Err string `json:"err,omitempty"`
}

type CategoryRequest struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	ExerciseIds []uint `json:"exercise_ids"`
}
type ExerciseRequest struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	CategoryIds []uint    `json:"category_ids"`
	Explain     []Explain `json:"explain"`
}

type CategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ExerciseResponse struct {
	ID         uint               `json:"id"`
	Name       string             `json:"name"`
	Explain    []Explain          `json:"explain"`
	Categories []CategoryResponse `json:"categories"`
}
type CategoryExerciseResponse struct {
	ID        uint                        `json:"id"`
	Name      string                      `json:"name"`
	Exercises []ExerciseCatregoryResponse `json:"exercises"`
}

type ExerciseCatregoryResponse struct {
	ID         uint               `json:"id"`
	Name       string             `json:"name"`
	Explain    []Explain          `json:"explain"`
	Categories []CategoryResponse `json:"categories"`
}

type MachineDto struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	MachineType uint   `json:"machine_type"`
	Memo        string `json:"memo"`
}

type PurposeDto struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type MeasureDto struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type RecommendRequest struct {
	ExerciseID   uint           `json:"exercise_id"`   // 운동아이디
	MachineIDs   []uint         `json:"machine_ids"`   // 기구아이디
	PurposeIDs   []uint         `json:"purpose_ids"`   // 목적아이디
	MeasureIds   []uint         `json:"measure_ids"`   // 측정항목 아이디
	IsAsymmetric bool           `json:"is_asymmetric"` // 비대칭 여부
	BodyType     uint           `json:"body_type"`     // 전신,상체,하체
	Locomotion   uint           `json:"locomotion"`
	IsGrip       bool           `json:"is_grip"`
	Afcs         []RecommendAfc `json:"afcs"`
}

type RecommendAfc struct {
	JointAction  uint          `json:"joint_action"`
	Rom          uint          `json:"rom"`
	ClinicDegree map[uint]uint `json:"clinic_degree"`
}

type Explain struct {
	Insert     interface{}            `json:"insert"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type RecommendResponse struct {
	Exercise     ExerciseResponse `json:"exercise"`
	Machines     []MachineDto     `json:"machines"`
	Purposes     []PurposeDto     `json:"purposes"`
	Measures     []MeasureDto     `json:"measure"`
	BodyType     uint             `json:"body_type"`
	IsAsymmetric bool             `json:"is_asymmetric"`
	Locomotion   uint             `json:"locomotion"`
	Afcs         []RecommendAfc   `json:"afcs"`
	IsGrip       *bool            `json:"is_grip"`
	// BodyRomClinicDegree map[uint]map[uint]map[uint]uint `json:"body_rom_clinic_degree,omitempty"`
}

type SearchRequest struct {
	Page uint   `form:"page"`
	Name string `form:"name"`
}
type BasicResponse struct {
	Msg string `json:"msg"`
}

// // for swagger ////
type SuccessResponse struct {
	Jwt string `json:"jwt"`
}
type ErrorResponse struct {
	Err string `json:"err"`
}
