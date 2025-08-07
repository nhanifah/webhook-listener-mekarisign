package model

import (
	"database/sql"

	"webhook-listener-mekarisign/database"
)

// Student merepresentasikan struktur tabel students di database
type Student struct {
	ID             string         `json:"id"`
	FullName       sql.NullString `json:"full_name"`
	Gender         sql.NullString `json:"gender"`
	WhatsAppNumber sql.NullString `json:"whatsapp_number"`
	Email          sql.NullString `json:"email"`
	NIK            sql.NullString `json:"nik"`
	Province       sql.NullString `json:"province"`
	City           sql.NullString `json:"city"`
	Subdistrict    sql.NullString `json:"subdistrict"`
	Village        sql.NullString `json:"village"`
	AddressDetail  sql.NullString `json:"address_detail"`
	LastEducation  sql.NullString `json:"last_education"`
	WorkNow        sql.NullString `json:"work_now"`
	ProgramType    sql.NullString `json:"program_type"`
	WantToWork     sql.NullString `json:"want_to_work"`
	BatchID        sql.NullString `json:"batch_id"`
	InterviewID    sql.NullString `json:"interview_id"`
	Dormitory      sql.NullString `json:"dormitory"`
	Installment    sql.NullString `json:"installment"`
	ReferralID     sql.NullString `json:"referral_id"`
	ReferralSource sql.NullString `json:"referral_source"`
	Birthdate      []uint8        `json:"birthdate"`
	GuardianStatus sql.NullString `json:"guardian_status"`
	GuardianName   sql.NullString `json:"guardian_name"`
	GuardianPhone  sql.NullString `json:"guardian_phone"`
	OtherInfo      sql.NullString `json:"other_info_exercise"`
	Progress       sql.NullString `json:"progress"`
	CreatedAt      []uint8        `json:"created_at"`
	UpdatedAt      []uint8        `json:"updated_at"`
	DeletedAt      []uint8        `json:"deleted_at"`
}

// GetAllStudents mengambil semua data mahasiswa dari database
func GetAllStudents() ([]Student, error) {
	rows, err := database.DB.Query(`SELECT 
		id, full_name, gender, whatsapp_number, email, nik, province, city, subdistrict, village, 
		address_detail, last_education, work_now, program_type, want_to_work, batch_id, interview_id, 
		dormitory, installment, referral_id, referral_source, birthdate, guardian_status, guardian_name, 
		guardian_phone, other_info_exercise, progress, created_at, updated_at, deleted_at 
		FROM students`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var s Student
		err := rows.Scan(
			&s.ID, &s.FullName, &s.Gender, &s.WhatsAppNumber, &s.Email, &s.NIK, &s.Province, &s.City,
			&s.Subdistrict, &s.Village, &s.AddressDetail, &s.LastEducation, &s.WorkNow, &s.ProgramType, &s.WantToWork,
			&s.BatchID, &s.InterviewID, &s.Dormitory, &s.Installment, &s.ReferralID, &s.ReferralSource, &s.Birthdate,
			&s.GuardianStatus, &s.GuardianName, &s.GuardianPhone, &s.OtherInfo, &s.Progress, &s.CreatedAt,
			&s.UpdatedAt, &s.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, nil
}

// GetStudentByID mengambil satu data mahasiswa berdasarkan ID
func GetStudentByID(id string) (*Student, error) {
	var s Student
	err := database.DB.QueryRow(`SELECT 
		id, full_name, gender, whatsapp_number, email, nik, province, city, subdistrict, village, 
		address_detail, last_education, work_now, program_type, want_to_work, batch_id, interview_id, 
		dormitory, installment, referral_id, referral_source, birthdate, guardian_status, guardian_name, 
		guardian_phone, other_info_exercise, progress, created_at, updated_at, deleted_at 
		FROM students WHERE id = ?`, id).Scan(
		&s.ID, &s.FullName, &s.Gender, &s.WhatsAppNumber, &s.Email, &s.NIK, &s.Province, &s.City,
		&s.Subdistrict, &s.Village, &s.AddressDetail, &s.LastEducation, &s.WorkNow, &s.ProgramType, &s.WantToWork,
		&s.BatchID, &s.InterviewID, &s.Dormitory, &s.Installment, &s.ReferralID, &s.ReferralSource, &s.Birthdate,
		&s.GuardianStatus, &s.GuardianName, &s.GuardianPhone, &s.OtherInfo, &s.Progress, &s.CreatedAt,
		&s.UpdatedAt, &s.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func GetStudentByEmail(email string) (*Student, error) {
	var s Student
	err := database.DB.QueryRow(`SELECT 
		id, full_name, gender, whatsapp_number, email, nik, province, city, subdistrict, village, 
		address_detail, last_education, work_now, program_type, want_to_work, batch_id, interview_id, 
		dormitory, installment, referral_id, referral_source, birthdate, guardian_status, guardian_name, 
		guardian_phone, other_info_exercise, progress, created_at, updated_at, deleted_at 
		FROM students WHERE email = ?`, email).Scan(
		&s.ID, &s.FullName, &s.Gender, &s.WhatsAppNumber, &s.Email, &s.NIK, &s.Province, &s.City,
		&s.Subdistrict, &s.Village, &s.AddressDetail, &s.LastEducation, &s.WorkNow, &s.ProgramType, &s.WantToWork,
		&s.BatchID, &s.InterviewID, &s.Dormitory, &s.Installment, &s.ReferralID, &s.ReferralSource, &s.Birthdate,
		&s.GuardianStatus, &s.GuardianName, &s.GuardianPhone, &s.OtherInfo, &s.Progress, &s.CreatedAt,
		&s.UpdatedAt, &s.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// CreateStudent menambahkan data mahasiswa baru ke database
func CreateStudent(s Student) (string, error) {
	result, err := database.DB.Exec(`INSERT INTO students (
		id, full_name, gender, whatsapp_number, email, nik, province, city, subdistrict, village, 
		address_detail, last_education, work_now, program_type, want_to_work, batch_id, interview_id, 
		dormitory, installment, referral_id, referral_source, birthdate, guardian_status, guardian_name, 
		guardian_phone, other_info_exercise, progress, created_at, updated_at, deleted_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.FullName, s.Gender, s.WhatsAppNumber, s.Email, s.NIK, s.Province, s.City,
		s.Subdistrict, s.Village, s.AddressDetail, s.LastEducation, s.WorkNow, s.ProgramType, s.WantToWork,
		s.BatchID, s.InterviewID, s.Dormitory, s.Installment, s.ReferralID, s.ReferralSource, s.Birthdate,
		s.GuardianStatus, s.GuardianName, s.GuardianPhone, s.OtherInfo, s.Progress, s.CreatedAt,
		s.UpdatedAt, s.DeletedAt,
	)
	if err != nil {
		return "", err
	}
	lastID, _ := result.LastInsertId()
	return string(lastID), nil
}

// UpdateStudent memperbarui data mahasiswa berdasarkan ID
func UpdateStudent(s Student) error {
	_, err := database.DB.Exec(`UPDATE students SET 
		full_name=?, gender=?, whatsapp_number=?, email=?, nik=?, province=?, city=?, subdistrict=?, village=?, 
		address_detail=?, last_education=?, work_now=?, program_type=?, want_to_work=?, batch_id=?, interview_id=?, 
		dormitory=?, installment=?, referral_id=?, referral_source=?, birthdate=?, guardian_status=?, guardian_name=?, 
		guardian_phone=?, other_info_exercise=?, progress=?, updated_at=? WHERE id=?`,
		s.FullName, s.Gender, s.WhatsAppNumber, s.Email, s.NIK, s.Province, s.City,
		s.Subdistrict, s.Village, s.AddressDetail, s.LastEducation, s.WorkNow, s.ProgramType, s.WantToWork,
		s.BatchID, s.InterviewID, s.Dormitory, s.Installment, s.ReferralID, s.ReferralSource, s.Birthdate,
		s.GuardianStatus, s.GuardianName, s.GuardianPhone, s.OtherInfo, s.Progress, s.UpdatedAt, s.ID,
	)
	return err
}

// DeleteStudent menghapus data mahasiswa berdasarkan ID
func DeleteStudent(id string) error {
	_, err := database.DB.Exec("DELETE FROM students WHERE id = ?", id)
	return err
}
