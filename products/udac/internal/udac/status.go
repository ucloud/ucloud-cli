package udac

const (
	productName    = "udac"
	resourceIDFlag = "udb-id" // list 命令的 --udb-id flag
	typeFlag       = "type"   // 实例类型 flag
)

// SupportedTypes 支持的实例类型列表
var SupportedTypes = []string{"mysql", "mongodb"}
