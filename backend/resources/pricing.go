package resources

import _ "embed"

// ModelPricing 随二进制携带同一份目录，独立归档安装也能使用核定的人民币价。
//
//go:embed model-pricing/model_prices_and_context_window.json
var ModelPricing []byte
