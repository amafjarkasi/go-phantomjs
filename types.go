package phantomjscloud

// UserRequest represents the root POST payload (IUserRequest)
type UserRequest struct {
	Pages []PageRequest `json:"pages"`
}

// PageRequest represents the IPageRequest interface
type PageRequest struct {
	URL             string          `json:"url"`
	Content         string          `json:"content,omitempty"`
	RenderType      string          `json:"renderType,omitempty"`
	OutputAsJson    bool            `json:"outputAsJson,omitempty"`
	OverseerScript  string          `json:"overseerScript,omitempty"`
	Proxy           string          `json:"proxy,omitempty"`
	Backend         string          `json:"backend,omitempty"`
	SuppressJson    []string        `json:"suppressJson,omitempty"`
	QueryJson       interface{}     `json:"queryJson,omitempty"`
	UrlSettings     *UrlSettings    `json:"urlSettings,omitempty"`
	Scripts         *Scripts        `json:"scripts,omitempty"`
	RequestSettings RequestSettings `json:"requestSettings,omitempty"`
	RenderSettings  RenderSettings  `json:"renderSettings,omitempty"`
}

type UrlSettings struct {
	Operation string            `json:"operation,omitempty"`
	Data      string            `json:"data,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Encoding  string            `json:"encoding,omitempty"`
}

type Scripts struct {
	Load         []string `json:"load,omitempty"`
	DomReady     []string `json:"domReady,omitempty"`
	LoadFinished []string `json:"loadFinished,omitempty"`
}

// RequestSettings represents the IRequestSettings interface
type RequestSettings struct {
	WaitInterval         int                `json:"waitInterval,omitempty"`
	IoWait               int                `json:"ioWait,omitempty"`
	MaxWait              int                `json:"maxWait,omitempty"`
	ResourceWait         int                `json:"resourceWait,omitempty"`
	ResourceTimeout      int                `json:"resourceTimeout,omitempty"`
	IgnoreImages         bool               `json:"ignoreImages,omitempty"`
	DisableJavascript    bool               `json:"disableJavascript,omitempty"`
	UserAgent            string             `json:"userAgent,omitempty"`
	DoneWhen             []DoneWhen         `json:"doneWhen,omitempty"`
	ResourceModifier     []ResourceModifier `json:"resourceModifier,omitempty"`
	RecordResourceBody   string             `json:"recordResourceBody,omitempty"`
	DisableSecureHeaders bool               `json:"disableSecureHeaders,omitempty"`
	WebSecurityEnabled   bool               `json:"webSecurityEnabled,omitempty"`
	XssAuditingEnabled   bool               `json:"xssAuditingEnabled,omitempty"`
	ClearCache           bool               `json:"clearCache,omitempty"`
	ClearCookies         bool               `json:"clearCookies,omitempty"`
	Cookies              []Cookie           `json:"cookies,omitempty"`
	DeleteCookies        []Cookie           `json:"deleteCookies,omitempty"`
	Authentication       *Authentication    `json:"authentication,omitempty"`
	CustomHeaders        map[string]string  `json:"customHeaders,omitempty"`
	EmulateDevice        string             `json:"emulateDevice,omitempty"`
	StopOnError          bool               `json:"stopOnError,omitempty"`
}

type Authentication struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	URL      string `json:"url,omitempty"`
	Path     string `json:"path,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
	Expires  int    `json:"expires,omitempty"`
}

type DoneWhen struct {
	Event      string `json:"event,omitempty"`
	Selector   string `json:"selector,omitempty"`
	Text       string `json:"text,omitempty"`
	StatusCode int    `json:"statusCode,omitempty"`
}

type ResourceModifier struct {
	Regex         string            `json:"regex,omitempty"`
	Type          string            `json:"type,omitempty"`
	IsBlacklisted bool              `json:"isBlacklisted,omitempty"`
	SetHeader     map[string]string `json:"setHeader,omitempty"`
	ChangeUrl     string            `json:"changeUrl,omitempty"`
}

// RenderSettings represents the IRenderSettings interface
type RenderSettings struct {
	Quality               int               `json:"quality,omitempty"`
	PassThroughHeaders    bool              `json:"passThroughHeaders,omitempty"`
	PassThroughStatusCode bool              `json:"passThroughStatusCode,omitempty"`
	Viewport              *Viewport         `json:"viewport,omitempty"`
	ClipRectangle         *ClipRectangle    `json:"clipRectangle,omitempty"`
	Selector              string            `json:"selector,omitempty"`
	ShadowDom             string            `json:"shadowDom,omitempty"`
	EmulateMedia          string            `json:"emulateMedia,omitempty"`
	ExtraResponseHeaders  map[string]string `json:"extraResponseHeaders,omitempty"`
	PdfOptions            *PdfOptions       `json:"pdfOptions,omitempty"`
	PngOptions            *PngOptions       `json:"pngOptions,omitempty"`
	IFrameMaxCount        int               `json:"iFrameMaxCount,omitempty"`
	IFrameMaxDepth        int               `json:"iFrameMaxDepth,omitempty"`
	OmitBackground        bool              `json:"omitBackground,omitempty"`
	RenderIFrame          bool              `json:"renderIFrame,omitempty"`
	ZoomFactor            float64           `json:"zoomFactor,omitempty"`
}

type PngOptions struct {
	CompressionLevel int `json:"compressionLevel,omitempty"`
}

type Viewport struct {
	Width             int     `json:"width"`
	Height            int     `json:"height"`
	DeviceScaleFactor float64 `json:"deviceScaleFactor,omitempty"`
	IsMobile          bool    `json:"isMobile,omitempty"`
	HasTouch          bool    `json:"hasTouch,omitempty"`
	IsLandscape       bool    `json:"isLandscape,omitempty"`
}

type ClipRectangle struct {
	Top    int `json:"top"`
	Left   int `json:"left"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type PdfOptions struct {
	Scale               float64 `json:"scale,omitempty"`
	DisplayHeaderFooter bool    `json:"displayHeaderFooter,omitempty"`
	HeaderTemplate      string  `json:"headerTemplate,omitempty"`
	FooterTemplate      string  `json:"footerTemplate,omitempty"`
	PrintBackground     bool    `json:"printBackground,omitempty"`
	Landscape           bool    `json:"landscape,omitempty"`
	PageRanges          string  `json:"pageRanges,omitempty"`
	Format              string  `json:"format,omitempty"`
	Width               string  `json:"width,omitempty"`
	Height              string  `json:"height,omitempty"`
	Margin              *Margin `json:"margin,omitempty"`
}

type Margin struct {
	Top    string `json:"top,omitempty"`
	Right  string `json:"right,omitempty"`
	Bottom string `json:"bottom,omitempty"`
	Left   string `json:"left,omitempty"`
}

// UserResponse represents the root response object
type UserResponse struct {
	PageResponses   []PageResponse         `json:"pageResponses"`
	Billing         Billing                `json:"billing"`
	Status          string                 `json:"status"`
	StatusMessage   string                 `json:"statusMessage,omitempty"`
	QueryJson       interface{}            `json:"queryJson,omitempty"`
	OriginalRequest interface{}            `json:"originalRequest,omitempty"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
}

type Billing struct {
	CreditCost float64 `json:"creditCost"`
	QuotaUsage float64 `json:"quotaUsage"`
}

// PageResponse represents the response object for each requested page
type PageResponse struct {
	Content          string                 `json:"content"`
	Metrics          Metrics                `json:"metrics"`
	Events           map[string][]Event     `json:"events"`
	StatusCode       int                    `json:"statusCode"`
	StatusText       string                 `json:"statusText"`
	Headers          map[string]string      `json:"headers"`
	Meta             map[string]interface{} `json:"meta"`
	DoneWhen         []DoneWhen             `json:"doneWhen,omitempty"`
	FrameData        *FrameData             `json:"frameData,omitempty"`
	Cookies          []Cookie               `json:"cookies,omitempty"`
	Errors           []string               `json:"errors,omitempty"`
	ContentErrors    []string               `json:"contentErrors,omitempty"`
	AutomationResult interface{}            `json:"automationResult,omitempty"`
	ScriptOutput     map[string]interface{} `json:"scriptOutput,omitempty"`
	EventPhase       string                 `json:"eventPhase,omitempty"`
	Resources        []interface{}          `json:"resources,omitempty"`
}

type Metrics struct {
	WaitInterval       int              `json:"waitInterval"`
	BillingRenderMs    int              `json:"billingRenderMs"`
	PageLoadStartTime  int              `json:"pageLoadStartTime"`
	PageLoadFinishTime int              `json:"pageLoadFinishTime"`
	TotalRenderTimeMs  int              `json:"totalRenderTimeMs"`
	ResourceSummary    *ResourceSummary `json:"resourceSummary,omitempty"`
}

type ResourceSummary struct {
	Aborted  int `json:"aborted,omitempty"`
	Active   int `json:"active,omitempty"`
	Complete int `json:"complete,omitempty"`
	Failed   int `json:"failed,omitempty"`
	Late     int `json:"late,omitempty"`
	Orphaned int `json:"orphaned,omitempty"`
}

type Event struct {
	URL        string            `json:"url,omitempty"`
	StatusCode int               `json:"statusCode,omitempty"`
	Method     string            `json:"method,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	PostData   string            `json:"postData,omitempty"`
	BodyBase64 string            `json:"bodyBase64,omitempty"`
}

type FrameData struct {
	Id          string       `json:"id"`
	Url         string       `json:"url"`
	Name        string       `json:"name"`
	Content     string       `json:"content,omitempty"`
	ChildFrames []*FrameData `json:"childFrames,omitempty"`
}

// Client Meta Response Headers
// PJSC Headers are returned in the HTTP Response Headers, which we'll parse.
type ResponseMetadata struct {
	BillingCreditCost float64
	ContentStatusCode int
	ContentDoneWhen   string
}
