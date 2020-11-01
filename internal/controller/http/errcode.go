package http

var codeText map[int]string

const (
	CodeSuccess      = 0
	CodeLackParam    = 5000 + iota // 缺少参数
	CodeInvalidParam               // 非法参数
	CodeAccessToken                // 获取access token 出错
	CodeVerifyToken                // 验证access token 出错
	CodeIllegalToken               // 非法token
	CodeNodeID                     // 获取 node id 失败
)

func init() {
	codeText = make(map[int]string)
	codeText[CodeSuccess] = "success"
	codeText[CodeLackParam] = "lack of param"
	codeText[CodeInvalidParam] = "invalid param"
	codeText[CodeVerifyToken] = "something wrong when verify token"
	codeText[CodeIllegalToken] = "illegal token"
	codeText[CodeNodeID] = "failed to get node id"
}
