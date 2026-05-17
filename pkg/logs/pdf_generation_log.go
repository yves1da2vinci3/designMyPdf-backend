package logs

import (
	"designmypdf/pkg/entities"
	"designmypdf/pkg/template"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"gorm.io/datatypes"
)

// FormatErrorWithBacktrace returns err text plus a stack trace when err is non-nil.
func FormatErrorWithBacktrace(err error) string {
	if err == nil {
		return ""
	}
	return err.Error() + "\n--- backtrace ---\n" + string(debug.Stack())
}

// RecordPdfGeneration persists a PDF generation log entry.
// When templateID is 0, templateUUID is used to resolve the template; if still unknown, no row is written.
func RecordPdfGeneration(keyID, templateID uint, templateUUID string, requestBody []byte, response map[string]interface{}, status entities.StatusCode, err error) error {
	if templateID == 0 && templateUUID != "" {
		templateSvc := template.NewService(template.Repository{})
		t, lookupErr := templateSvc.GetByUUID(templateUUID)
		if lookupErr != nil {
			log.Printf("logs: skip PDF log (template %q not found): %v", templateUUID, lookupErr)
			return lookupErr
		}
		templateID = t.ID
	}
	if templateID == 0 {
		log.Printf("logs: skip PDF log (no template id for key %d)", keyID)
		return errors.New("template id required for PDF log")
	}

	responseBodyBytes, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		return fmt.Errorf("marshal response body: %w", marshalErr)
	}

	logEntry := &entities.Log{
		TemplateID:   templateID,
		KeyID:        keyID,
		CalledAt:     time.Now(),
		RequestBody:  datatypes.JSON(requestBody),
		ResponseBody: datatypes.JSON(responseBodyBytes),
		StatusCode:   status,
		ErrorMessage: FormatErrorWithBacktrace(err),
	}

	logService := NewService(Repository{})
	if createErr := logService.CreateLog(logEntry); createErr != nil {
		log.Printf("logs: failed to create PDF generation log: %v", createErr)
		return createErr
	}
	return nil
}
