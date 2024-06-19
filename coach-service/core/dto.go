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
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type CategoryResponse struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	Exercises []ExerciseResponse `json:"exercises"`
}

type ExerciseRequest struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	CategoryId uint   `json:"category_id"`
}
type ExerciseResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	BodyType uint   `json:"body_type"`
}

type MachineDto struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type PurposeDto struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type RecommendRequest struct {
	ExerciseID          uint                            `json:"exercise_id"` // 운동아이디
	MachineIDs          []uint                          `json:"machine_ids"` // 기구아이디
	PurposeIDs          []uint                          `json:"purpose_ids"` // 목적아이디
	Asymmetric          bool                            `json:"asymmetric"`  // 비대칭 여부
	BodyType            uint                            `json:"body_type"`
	BodyRomClinicDegree map[uint]map[uint]map[uint]uint `json:"body_rom_clinic_degree"` // 증상id : 정도
	TrRom               uint                            `json:"tr_rom"`
	Locomotion          uint                            `json:"locomotion"`
}

type RecommendResponse struct {
	Category            CategoryRequest  `json:"category"`
	Exercise            ExerciseResponse `json:"exercise"`
	Machines            []MachineDto     `json:"machines"`
	Purposes            []PurposeDto     `json:"purposes"`
	Asymmetric          bool             `json:"asymmetric"`
	TrRom               uint             `json:"tr_rom"`
	Locomotion          uint             `json:"locomotion"`
	BodyRomClinicDegree map[uint]map[uint]map[uint]uint
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
