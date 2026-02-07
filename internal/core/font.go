package core

import _ "embed"

//go:embed assets/fonts/NotoSansCJKsc-Regular.otf
var defaultFontData []byte

const defaultFontName = "Noto Sans CJK SC"
