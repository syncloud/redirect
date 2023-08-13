package user

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
)

const StatusFile = "/var/www/redirect/user.cleanup.last"

type CleanerState struct {
	file   string
	logger *zap.Logger
}

func NewCleanerState(logger *zap.Logger) *CleanerState {
	return &CleanerState{
		file:   StatusFile,
		logger: logger,
	}
}

func (s *CleanerState) Get() (int64, error) {
	if _, err := os.Stat(s.file); err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("does not exist yet, will use 0", zap.String("file", s.file))
			return 0, nil
		} else {
			return 0, err
		}
	}

	contents, err := os.ReadFile(s.file)
	if err != nil {
		return 0, err
	}
	userIdStr := strings.TrimSpace(string(contents))
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func (s *CleanerState) Set(userId int64) error {
	return os.WriteFile(s.file, []byte(fmt.Sprintf("%d", userId)), 0644)
}
