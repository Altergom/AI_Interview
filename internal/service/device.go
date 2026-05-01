package service

import "context"

// DeviceService 设备检测模块业务逻辑接口
type DeviceService interface {
	// Check 记录设备信息，返回检测结果。
	Check(ctx context.Context, req DeviceCheckRequest) (*DeviceCheckResult, error)
}

type DeviceCheckRequest struct {
	UserID        string
	HasMicrophone bool
	HasCamera     bool
	Browser       string
	OS            string
}

type DeviceCheckResult struct {
	Status  string
	Message string
}
