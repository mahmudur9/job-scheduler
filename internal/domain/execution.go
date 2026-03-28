package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

func GenerateExecutionKey(jobID uuid.UUID, scheduleTime int64) string {
	data := fmt.Sprintf("%s:%d", jobID, scheduleTime)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
