// Package taskcode provides task error codes. 与alarm-compute同步
package taskcode

// TaskStatusErr 任务执行状态错误
type TaskStatusErr struct {
	ErrCode   int32
	ErrName   string
	ErrDetail string
}

var (
	PointSvcErr TaskStatusErr = TaskStatusErr{1, "数据服务异常", ""}

	PointDataLackErr TaskStatusErr = TaskStatusErr{2, "测点数据缺失", ""}
	ExprAnalyzeErr   TaskStatusErr = TaskStatusErr{3, "表达式计算失败", ""}

	// 恢复分析错误
	RestoreAnalyzeErr TaskStatusErr = TaskStatusErr{4, "告警恢复任务计算失败", ""}
	UnKnownErr        TaskStatusErr = TaskStatusErr{5, "未知错误", ""}
)

// 实现 error 接口
func (e *TaskStatusErr) Error() string {
	return e.ErrName
}

// GetErrCode 获取错误码
func (e *TaskStatusErr) GetErrCode() int32 {
	return e.ErrCode
}

// GetErrMsg 获取错误信息
func (e *TaskStatusErr) GetErrMsg() string {
	return e.ErrName
}

// GetErrDetail 获取错误详情
func (e *TaskStatusErr) GetErrDetail() string {
	return e.ErrDetail
}

// JudgeErrType JudgeErrType
func (e *TaskStatusErr) JudgeErrType(template *TaskStatusErr) bool {
	return e.ErrCode == template.ErrCode
}

// NewErr NewErr
func NewErr(errType *TaskStatusErr, detail string) (err *TaskStatusErr) {
	return &TaskStatusErr{errType.ErrCode, errType.ErrName, detail}
}
