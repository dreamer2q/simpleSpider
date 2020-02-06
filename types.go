package spider

type URL []string

type Domain []string

type ContentType []string

const (
	TextAll   = "text/*"
	HtmlType  = "text/html"
	CssType   = "text/css"
	PlainType = "text/plain"

	ApplicationAll = "application/*"
	JsType         = "application/javascript"

	ImageAll = "image/*"
	PngType  = "image/png"
	IconType = "image/x-icon"

	VideoAll = "video/*"
	Mp4Type  = "video/mp4"
)
