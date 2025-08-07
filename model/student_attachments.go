package model

import (
	"database/sql"

	"errors"

	"time"

	"go.uber.org/zap"

	"webhook-listener-mekarisign/database"
	"webhook-listener-mekarisign/logger"
)

type StudentAttachment struct {
	ID         string         `json:"id"`
	StudentID  sql.NullString `json:"student_id"`
	FileName   string         `json:"file_name"`
	FileURL    string         `json:"file_url"`
	UploadedAt sql.NullTime   `json:"uploaded_at"`
	CreatedAt  sql.NullTime   `json:"created_at"`
	UpdatedAt  sql.NullTime   `json:"updated_at"`
	DeletedAt  sql.NullTime   `json:"deleted_at"`
}

// GetAllStudentAttachments mengambil semua data lampiran mahasiswa dari database
func GetAllStudentAttachments() ([]StudentAttachment, error) {
	rows, err := database.DB.Query(`SELECT 
		id, student_id, file_name, file_url, uploaded_at, created_at, updated_at, deleted_at 
		FROM student_attachments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentAttachments []StudentAttachment
	for rows.Next() {
		var sa StudentAttachment
		err := rows.Scan(
			&sa.ID, &sa.StudentID, &sa.FileName, &sa.FileURL, &sa.UploadedAt, &sa.CreatedAt, &sa.UpdatedAt, &sa.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		studentAttachments = append(studentAttachments, sa)
	}

	return studentAttachments, nil
}

// GetStudentAttachmentByID mengambil data lampiran mahasiswa berdasarkan ID dari database
func GetStudentAttachmentByID(id string) (StudentAttachment, error) {
	var sa StudentAttachment
	err := database.DB.QueryRow(`SELECT 
		id, student_id, file_name, file_url, uploaded_at, created_at, updated_at, deleted_at 
		FROM student_attachments WHERE id = ?`, id).Scan(
		&sa.ID, &sa.StudentID, &sa.FileName, &sa.FileURL, &sa.UploadedAt, &sa.CreatedAt, &sa.UpdatedAt, &sa.DeletedAt,
	)
	if err != nil {
		return sa, err
	}

	return sa, nil
}

// GetStudentAttachmentByStudentID mengambil data lampiran mahasiswa berdasarkan ID mahasiswa dari database
func GetStudentAttachmentByStudentID(studentID string) ([]StudentAttachment, error) {
	rows, err := database.DB.Query(`SELECT 
		id, student_id, file_name, file_url, uploaded_at, created_at, updated_at, deleted_at 
		FROM student_attachments WHERE student_id = ?`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentAttachments []StudentAttachment
	for rows.Next() {
		var sa StudentAttachment
		err := rows.Scan(
			&sa.ID, &sa.StudentID, &sa.FileName, &sa.FileURL, &sa.UploadedAt, &sa.CreatedAt, &sa.UpdatedAt, &sa.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		studentAttachments = append(studentAttachments, sa)
	}

	return studentAttachments, nil
}

// CreateStudentAttachment membuat data lampiran mahasiswa baru di database
func CreateStudentAttachment(sa StudentAttachment) (string, error) {
	result, err := database.DB.Exec(`INSERT INTO student_attachments 
		(id, student_id, file_name, file_url, uploaded_at, created_at, updated_at, deleted_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, sa.ID, sa.StudentID, sa.FileName, sa.FileURL, sa.UploadedAt, sa.CreatedAt, sa.UpdatedAt, sa.DeletedAt)
	if err != nil {
		return "", err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	return string(id), nil
}

// UpdateStudentAttachment mengubah data lampiran mahasiswa di database
func UpdateStudentAttachment(sa StudentAttachment) error {
	logger.NewLogger().Info("UpdateStudentAttachment", zap.Any("student_attachment", sa))
	_, err := database.DB.Exec(`UPDATE student_attachments SET 
		student_id = ?, file_name = ?, file_url = ?, uploaded_at = ?, created_at = ?, updated_at = ?, deleted_at = ? 
		WHERE id = ?`, sa.StudentID, sa.FileName, sa.FileURL, sa.UploadedAt, sa.CreatedAt, sa.UpdatedAt, sa.DeletedAt, sa.ID)
	if err != nil {
		return err
	}

	return nil
}

// create or update student attachment
func CreateOrUpdateStudentAttachment(sa StudentAttachment) (string, error) {
	existing, err := GetStudentAttachmentByID(sa.ID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// data belum ada, silakan create
			return CreateStudentAttachment(sa)
		}
		// ada error lain, kembalikan errornya aja
		return "", err
	}

	// data ditemukan, lakukan update
	sa.UploadedAt = existing.UploadedAt
	sa.CreatedAt = existing.CreatedAt
	sa.UpdatedAt = sql.NullTime{Time: time.Now(), Valid: true}

	// logger.NewLogger().Info("UpdateStudentAttachment", zap.Any("student_attachment", sa))

	err = UpdateStudentAttachment(sa)
	if err != nil {
		return "", err
	}
	return sa.ID, nil
}
