package constant

// ============================================================
// 业务响应码
// ============================================================

const (
	CodeOK int32 = 200

	// --- 通用客户端错误 4xx ---
	CodeInvalidParams int32 = 400
	CodeUnauthorized  int32 = 401

	// --- 服务端错误 6xx ---
	CodeInternal int32 = 600

	// --- RPC / 异步错误 ---
	CodeRPCError int32 = 1001
)
