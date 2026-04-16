// This file defines the shared bridge contracts and protobuf-wire codec used
// by Lina dynamic plugin runtime execution.

package pluginbridge

import (
	"encoding/base64"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"google.golang.org/protobuf/encoding/protowire"
)

const (
	// CodecProtobuf is the only supported executable bridge envelope codec.
	CodecProtobuf = "protobuf"

	// AccessPublic allows anonymous access.
	AccessPublic = "public"
	// AccessLogin requires authenticated access.
	AccessLogin = "login"

	// RuntimeKindWasm identifies a wasm runtime artifact.
	RuntimeKindWasm = "wasm"
	// ABIVersionV1 is the current bridge ABI version.
	ABIVersionV1 = "v1"
	// SupportedABIVersion is the current runtime artifact ABI version.
	SupportedABIVersion = ABIVersionV1

	// DefaultGuestAllocExport is the default guest allocator export.
	DefaultGuestAllocExport = "lina_dynamic_route_alloc"
	// DefaultGuestExecuteExport is the default guest executor export.
	DefaultGuestExecuteExport = "lina_dynamic_route_execute"
)

const (
	bridgeFailureCodeUnauthorized = "UNAUTHORIZED"
	bridgeFailureCodeForbidden    = "FORBIDDEN"
	bridgeFailureCodeBadRequest   = "BAD_REQUEST"
	bridgeFailureCodeNotFound     = "NOT_FOUND"
	bridgeFailureCodeInternal     = "INTERNAL_ERROR"
)

// validOperLogTags defines all valid semantic operLog tag values for route contracts.
var validOperLogTags = map[string]struct{}{
	"create": {},
	"update": {},
	"delete": {},
	"export": {},
	"import": {},
	"other":  {},
}

// validOperLogTagList lists all valid operLog tags for error messages.
var validOperLogTagList = []string{"create", "update", "delete", "export", "import", "other"}

// RouteContract describes one dynamic plugin route contract embedded into the artifact.
type RouteContract struct {
	Path        string   `json:"path" yaml:"path"`
	Method      string   `json:"method" yaml:"method"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string   `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Access      string   `json:"access,omitempty" yaml:"access,omitempty"`
	Permission  string   `json:"permission,omitempty" yaml:"permission,omitempty"`
	OperLog     string   `json:"operLog,omitempty" yaml:"operLog,omitempty"`
	RequestType string   `json:"requestType,omitempty" yaml:"requestType,omitempty"`
}

// BridgeSpec defines the stable guest ABI contract embedded into the artifact.
type BridgeSpec struct {
	ABIVersion     string `json:"abiVersion" yaml:"abiVersion"`
	RuntimeKind    string `json:"runtimeKind" yaml:"runtimeKind"`
	RouteExecution bool   `json:"routeExecution" yaml:"routeExecution"`
	RequestCodec   string `json:"requestCodec,omitempty" yaml:"requestCodec,omitempty"`
	ResponseCodec  string `json:"responseCodec,omitempty" yaml:"responseCodec,omitempty"`
	AllocExport    string `json:"allocExport,omitempty" yaml:"allocExport,omitempty"`
	ExecuteExport  string `json:"executeExport,omitempty" yaml:"executeExport,omitempty"`
}

// BridgeRequestEnvelopeV1 is the host-to-guest request envelope.
type BridgeRequestEnvelopeV1 struct {
	PluginID  string                 `json:"pluginId"`
	Route     *RouteMatchSnapshotV1  `json:"route,omitempty"`
	Request   *HTTPRequestSnapshotV1 `json:"request,omitempty"`
	Identity  *IdentitySnapshotV1    `json:"identity,omitempty"`
	RequestID string                 `json:"requestId,omitempty"`
}

// RouteMatchSnapshotV1 describes the matched route and host path mapping.
type RouteMatchSnapshotV1 struct {
	Method       string              `json:"method,omitempty"`
	PublicPath   string              `json:"publicPath,omitempty"`
	InternalPath string              `json:"internalPath,omitempty"`
	RoutePath    string              `json:"routePath,omitempty"`
	Access       string              `json:"access,omitempty"`
	Permission   string              `json:"permission,omitempty"`
	RequestType  string              `json:"requestType,omitempty"`
	PathParams   map[string]string   `json:"pathParams,omitempty"`
	QueryValues  map[string][]string `json:"queryValues,omitempty"`
}

// HTTPRequestSnapshotV1 describes the sanitized inbound HTTP request.
type HTTPRequestSnapshotV1 struct {
	Method       string              `json:"method,omitempty"`
	PublicPath   string              `json:"publicPath,omitempty"`
	InternalPath string              `json:"internalPath,omitempty"`
	RawPath      string              `json:"rawPath,omitempty"`
	RawQuery     string              `json:"rawQuery,omitempty"`
	Host         string              `json:"host,omitempty"`
	Scheme       string              `json:"scheme,omitempty"`
	RemoteAddr   string              `json:"remoteAddr,omitempty"`
	ClientIP     string              `json:"clientIp,omitempty"`
	ContentType  string              `json:"contentType,omitempty"`
	Headers      map[string][]string `json:"headers,omitempty"`
	Cookies      map[string]string   `json:"cookies,omitempty"`
	Body         []byte              `json:"body,omitempty"`
}

// IdentitySnapshotV1 describes authenticated user context injected by the host.
type IdentitySnapshotV1 struct {
	TokenID      string   `json:"tokenId,omitempty"`
	UserID       int32    `json:"userId,omitempty"`
	Username     string   `json:"username,omitempty"`
	Status       int32    `json:"status,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	RoleNames    []string `json:"roleNames,omitempty"`
	IsSuperAdmin bool     `json:"isSuperAdmin,omitempty"`
}

// BridgeResponseEnvelopeV1 is the guest-to-host response envelope.
type BridgeResponseEnvelopeV1 struct {
	StatusCode  int32               `json:"statusCode,omitempty"`
	ContentType string              `json:"contentType,omitempty"`
	Headers     map[string][]string `json:"headers,omitempty"`
	Body        []byte              `json:"body,omitempty"`
	Failure     *BridgeFailureV1    `json:"failure,omitempty"`
}

// BridgeFailureV1 contains normalized execution failure metadata.
type BridgeFailureV1 struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// ValidateRouteContracts validates one plugin's route declarations in-place.
func ValidateRouteContracts(pluginID string, routes []*RouteContract) error {
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		if route == nil {
			return gerror.New("动态路由合同不能为空")
		}
		normalizeRouteContract(route)
		if route.Path == "" {
			return gerror.New("动态路由 path 不能为空")
		}
		if !strings.HasPrefix(route.Path, "/") {
			return gerror.Newf("动态路由 path 必须以 / 开头: %s", route.Path)
		}
		if route.Method == "" {
			return gerror.Newf("动态路由 method 不能为空: %s", route.Path)
		}
		switch route.Access {
		case "", AccessLogin:
			route.Access = AccessLogin
		case AccessPublic:
		default:
			return gerror.Newf("动态路由 access 仅支持 public/login: %s %s", route.Method, route.Path)
		}
		if route.Access == AccessPublic {
			if route.Permission != "" {
				return gerror.Newf("public 动态路由不能声明 permission: %s %s", route.Method, route.Path)
			}
			if route.OperLog != "" {
				return gerror.Newf("public 动态路由不能声明 operLog: %s %s", route.Method, route.Path)
			}
		}
		if route.Permission != "" {
			parts := strings.Split(route.Permission, ":")
			if len(parts) != 3 {
				return gerror.Newf("动态路由 permission 必须使用 {pluginId}:{resource}:{action} 格式: %s", route.Permission)
			}
			if strings.TrimSpace(parts[0]) != strings.TrimSpace(pluginID) {
				return gerror.Newf("动态路由 permission 必须以前缀 %s: 开头: %s", pluginID, route.Permission)
			}
			if strings.TrimSpace(parts[1]) == "" || strings.TrimSpace(parts[2]) == "" {
				return gerror.Newf("动态路由 permission 资源与动作不能为空: %s", route.Permission)
			}
		}
		if route.OperLog != "" {
			if _, ok := validOperLogTags[route.OperLog]; !ok {
				return gerror.Newf("动态路由 operLog 仅支持 %v: %s %s", validOperLogTagList, route.Method, route.Path)
			}
		}
		key := route.Method + " " + route.Path
		if _, ok := seen[key]; ok {
			return gerror.Newf("动态路由 method + path 不能重复: %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// NormalizeBridgeSpec normalizes bridge defaults in-place.
func NormalizeBridgeSpec(spec *BridgeSpec) {
	if spec == nil {
		return
	}
	spec.ABIVersion = normalizeLower(spec.ABIVersion, ABIVersionV1)
	spec.RuntimeKind = normalizeLower(spec.RuntimeKind, RuntimeKindWasm)
	spec.RequestCodec = normalizeLower(spec.RequestCodec, "")
	spec.ResponseCodec = normalizeLower(spec.ResponseCodec, "")
	spec.AllocExport = strings.TrimSpace(spec.AllocExport)
	spec.ExecuteExport = strings.TrimSpace(spec.ExecuteExport)
	if spec.AllocExport == "" {
		spec.AllocExport = DefaultGuestAllocExport
	}
	if spec.ExecuteExport == "" {
		spec.ExecuteExport = DefaultGuestExecuteExport
	}
}

// ValidateBridgeSpec validates bridge ABI compatibility in-place.
func ValidateBridgeSpec(spec *BridgeSpec) error {
	if spec == nil {
		return nil
	}
	NormalizeBridgeSpec(spec)
	if spec.ABIVersion != ABIVersionV1 {
		return gerror.Newf("动态路由 bridge ABI 版本不支持: %s", spec.ABIVersion)
	}
	if spec.RuntimeKind != RuntimeKindWasm {
		return gerror.Newf("动态路由 bridge runtimeKind 仅支持 wasm: %s", spec.RuntimeKind)
	}
	if !spec.RouteExecution {
		return nil
	}
	if spec.RequestCodec != CodecProtobuf || spec.ResponseCodec != CodecProtobuf {
		return gerror.Newf(
			"动态路由 bridge 可执行模式仅支持 protobuf 编解码: request=%s response=%s",
			spec.RequestCodec,
			spec.ResponseCodec,
		)
	}
	if spec.AllocExport == "" || spec.ExecuteExport == "" {
		return gerror.New("动态路由 bridge 可执行模式缺少 guest 导出函数")
	}
	return nil
}

// EncodeRequestEnvelope encodes one request envelope into protobuf wire bytes.
func EncodeRequestEnvelope(in *BridgeRequestEnvelopeV1) ([]byte, error) {
	if in == nil {
		return nil, gerror.New("bridge request envelope 不能为空")
	}
	return marshalRequestEnvelope(in), nil
}

// DecodeRequestEnvelope decodes one request envelope from protobuf wire bytes.
func DecodeRequestEnvelope(content []byte) (*BridgeRequestEnvelopeV1, error) {
	out := &BridgeRequestEnvelopeV1{}
	if err := unmarshalRequestEnvelope(content, out); err != nil {
		return nil, err
	}
	return out, nil
}

// EncodeResponseEnvelope encodes one response envelope into protobuf wire bytes.
func EncodeResponseEnvelope(in *BridgeResponseEnvelopeV1) ([]byte, error) {
	if in == nil {
		return nil, gerror.New("bridge response envelope 不能为空")
	}
	return marshalResponseEnvelope(in), nil
}

// DecodeResponseEnvelope decodes one response envelope from protobuf wire bytes.
func DecodeResponseEnvelope(content []byte) (*BridgeResponseEnvelopeV1, error) {
	out := &BridgeResponseEnvelopeV1{}
	if err := unmarshalResponseEnvelope(content, out); err != nil {
		return nil, err
	}
	return out, nil
}

// NewSuccessResponse builds one normalized bridge success response.
func NewSuccessResponse(statusCode int, contentType string, body []byte) *BridgeResponseEnvelopeV1 {
	return &BridgeResponseEnvelopeV1{
		StatusCode:  int32(statusCode),
		ContentType: strings.TrimSpace(contentType),
		Body:        append([]byte(nil), body...),
	}
}

// NewJSONResponse builds one JSON response using the provided raw bytes.
func NewJSONResponse(statusCode int, body []byte) *BridgeResponseEnvelopeV1 {
	return NewSuccessResponse(statusCode, "application/json", body)
}

// NewFailureResponse builds one normalized failure response with a plain-text body.
func NewFailureResponse(statusCode int, code string, message string) *BridgeResponseEnvelopeV1 {
	content := strings.TrimSpace(message)
	response := &BridgeResponseEnvelopeV1{
		StatusCode:  int32(statusCode),
		ContentType: "text/plain; charset=utf-8",
		Body:        []byte(content),
		Failure: &BridgeFailureV1{
			Code:    strings.TrimSpace(code),
			Message: content,
		},
	}
	return response
}

// NewUnauthorizedResponse builds a normalized 401 response.
func NewUnauthorizedResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(401, bridgeFailureCodeUnauthorized, messageOrDefault(message, "Unauthorized"))
}

// NewForbiddenResponse builds a normalized 403 response.
func NewForbiddenResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(403, bridgeFailureCodeForbidden, messageOrDefault(message, "Forbidden"))
}

// NewBadRequestResponse builds a normalized 400 response.
func NewBadRequestResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(400, bridgeFailureCodeBadRequest, messageOrDefault(message, "Bad Request"))
}

// NewNotFoundResponse builds a normalized 404 response.
func NewNotFoundResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(404, bridgeFailureCodeNotFound, messageOrDefault(message, "Not Found"))
}

// NewInternalErrorResponse builds a normalized 500 response.
func NewInternalErrorResponse(message string) *BridgeResponseEnvelopeV1 {
	return NewFailureResponse(500, bridgeFailureCodeInternal, messageOrDefault(message, "Internal Server Error"))
}

func messageOrDefault(value string, fallback string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return fallback
	}
	return normalized
}

func normalizeRouteContract(route *RouteContract) {
	route.Path = strings.TrimSpace(route.Path)
	route.Method = strings.ToUpper(strings.TrimSpace(route.Method))
	route.Summary = strings.TrimSpace(route.Summary)
	route.Description = strings.TrimSpace(route.Description)
	route.Access = strings.ToLower(strings.TrimSpace(route.Access))
	route.Permission = strings.TrimSpace(route.Permission)
	route.OperLog = strings.ToLower(strings.TrimSpace(route.OperLog))
	route.RequestType = strings.TrimSpace(route.RequestType)
	if len(route.Tags) > 0 {
		tags := make([]string, 0, len(route.Tags))
		for _, item := range route.Tags {
			normalized := strings.TrimSpace(item)
			if normalized == "" {
				continue
			}
			tags = append(tags, normalized)
		}
		route.Tags = tags
	}
}

func normalizeLower(value string, defaultValue string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return defaultValue
	}
	return normalized
}

func marshalRequestEnvelope(in *BridgeRequestEnvelopeV1) []byte {
	var content []byte
	if value := strings.TrimSpace(in.PluginID); value != "" {
		content = appendStringField(content, 1, value)
	}
	if in.Route != nil {
		content = appendBytesField(content, 2, marshalRouteSnapshot(in.Route))
	}
	if in.Request != nil {
		content = appendBytesField(content, 3, marshalRequestSnapshot(in.Request))
	}
	if in.Identity != nil {
		content = appendBytesField(content, 4, marshalIdentitySnapshot(in.Identity))
	}
	if value := strings.TrimSpace(in.RequestID); value != "" {
		content = appendStringField(content, 5, value)
	}
	return content
}

func unmarshalRequestEnvelope(content []byte, out *BridgeRequestEnvelopeV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("解析 bridge request tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 bridge request pluginId 失败")
			}
			out.PluginID = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 bridge request route 失败")
			}
			out.Route = &RouteMatchSnapshotV1{}
			if err := unmarshalRouteSnapshot(value, out.Route); err != nil {
				return err
			}
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 bridge request request 失败")
			}
			out.Request = &HTTPRequestSnapshotV1{}
			if err := unmarshalRequestSnapshot(value, out.Request); err != nil {
				return err
			}
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 bridge request identity 失败")
			}
			out.Identity = &IdentitySnapshotV1{}
			if err := unmarshalIdentitySnapshot(value, out.Identity); err != nil {
				return err
			}
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 bridge request requestId 失败")
			}
			out.RequestID = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("跳过未知 bridge request 字段失败")
			}
			content = content[size:]
		}
	}
	return nil
}

func marshalResponseEnvelope(in *BridgeResponseEnvelopeV1) []byte {
	var content []byte
	if in.StatusCode != 0 {
		content = appendVarintField(content, 1, uint64(in.StatusCode))
	}
	if value := strings.TrimSpace(in.ContentType); value != "" {
		content = appendStringField(content, 2, value)
	}
	if len(in.Headers) > 0 {
		content = appendHeaderMap(content, 3, in.Headers)
	}
	if len(in.Body) > 0 {
		content = appendBytesField(content, 4, append([]byte(nil), in.Body...))
	}
	if in.Failure != nil {
		content = appendBytesField(content, 5, marshalFailure(in.Failure))
	}
	return content
}

func unmarshalResponseEnvelope(content []byte, out *BridgeResponseEnvelopeV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("解析 bridge response tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("解析 bridge response statusCode 失败")
			}
			out.StatusCode = int32(value)
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 bridge response contentType 失败")
			}
			out.ContentType = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 bridge response headers 失败")
			}
			if out.Headers == nil {
				out.Headers = make(map[string][]string)
			}
			if err := unmarshalHeaderEntry(value, out.Headers); err != nil {
				return err
			}
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 bridge response body 失败")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 bridge response failure 失败")
			}
			out.Failure = &BridgeFailureV1{}
			if err := unmarshalFailure(value, out.Failure); err != nil {
				return err
			}
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("跳过未知 bridge response 字段失败")
			}
			content = content[size:]
		}
	}
	return nil
}

func marshalRouteSnapshot(in *RouteMatchSnapshotV1) []byte {
	var content []byte
	if value := strings.TrimSpace(in.Method); value != "" {
		content = appendStringField(content, 1, value)
	}
	if value := strings.TrimSpace(in.PublicPath); value != "" {
		content = appendStringField(content, 2, value)
	}
	if value := strings.TrimSpace(in.InternalPath); value != "" {
		content = appendStringField(content, 3, value)
	}
	if value := strings.TrimSpace(in.RoutePath); value != "" {
		content = appendStringField(content, 4, value)
	}
	if value := strings.TrimSpace(in.Access); value != "" {
		content = appendStringField(content, 5, value)
	}
	if value := strings.TrimSpace(in.Permission); value != "" {
		content = appendStringField(content, 6, value)
	}
	if value := strings.TrimSpace(in.RequestType); value != "" {
		content = appendStringField(content, 7, value)
	}
	if len(in.PathParams) > 0 {
		content = appendStringMap(content, 8, in.PathParams)
	}
	if len(in.QueryValues) > 0 {
		content = appendStringListMap(content, 9, in.QueryValues)
	}
	return content
}

func unmarshalRouteSnapshot(content []byte, out *RouteMatchSnapshotV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("解析 route snapshot tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 route snapshot method 失败")
			}
			out.Method = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 route snapshot publicPath 失败")
			}
			out.PublicPath = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 route snapshot internalPath 失败")
			}
			out.InternalPath = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 route snapshot routePath 失败")
			}
			out.RoutePath = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 route snapshot access 失败")
			}
			out.Access = value
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 route snapshot permission 失败")
			}
			out.Permission = value
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 route snapshot requestType 失败")
			}
			out.RequestType = value
			content = content[size:]
		case 8:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 route snapshot pathParams 失败")
			}
			if out.PathParams == nil {
				out.PathParams = make(map[string]string)
			}
			if err := unmarshalStringEntry(value, out.PathParams); err != nil {
				return err
			}
			content = content[size:]
		case 9:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 route snapshot queryValues 失败")
			}
			if out.QueryValues == nil {
				out.QueryValues = make(map[string][]string)
			}
			if err := unmarshalStringListEntry(value, out.QueryValues); err != nil {
				return err
			}
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("跳过未知 route snapshot 字段失败")
			}
			content = content[size:]
		}
	}
	return nil
}

func marshalRequestSnapshot(in *HTTPRequestSnapshotV1) []byte {
	var content []byte
	appendStringField := func(fieldNumber protowire.Number, value string) {
		if normalized := strings.TrimSpace(value); normalized != "" {
			content = appendStringFieldContent(content, fieldNumber, normalized)
		}
	}
	appendStringField(1, in.Method)
	appendStringField(2, in.PublicPath)
	appendStringField(3, in.InternalPath)
	appendStringField(4, in.RawPath)
	appendStringField(5, in.RawQuery)
	appendStringField(6, in.Host)
	appendStringField(7, in.Scheme)
	appendStringField(8, in.RemoteAddr)
	appendStringField(9, in.ClientIP)
	appendStringField(10, in.ContentType)
	if len(in.Headers) > 0 {
		content = appendHeaderMap(content, 11, in.Headers)
	}
	if len(in.Cookies) > 0 {
		content = appendStringMap(content, 12, in.Cookies)
	}
	if len(in.Body) > 0 {
		content = appendBytesField(content, 13, append([]byte(nil), in.Body...))
	}
	return content
}

func unmarshalRequestSnapshot(content []byte, out *HTTPRequestSnapshotV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("解析 request snapshot tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot method 失败")
			}
			out.Method = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot publicPath 失败")
			}
			out.PublicPath = value
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot internalPath 失败")
			}
			out.InternalPath = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot rawPath 失败")
			}
			out.RawPath = value
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot rawQuery 失败")
			}
			out.RawQuery = value
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot host 失败")
			}
			out.Host = value
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot scheme 失败")
			}
			out.Scheme = value
			content = content[size:]
		case 8:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot remoteAddr 失败")
			}
			out.RemoteAddr = value
			content = content[size:]
		case 9:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot clientIp 失败")
			}
			out.ClientIP = value
			content = content[size:]
		case 10:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 request snapshot contentType 失败")
			}
			out.ContentType = value
			content = content[size:]
		case 11:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 request snapshot headers 失败")
			}
			if out.Headers == nil {
				out.Headers = make(map[string][]string)
			}
			if err := unmarshalHeaderEntry(value, out.Headers); err != nil {
				return err
			}
			content = content[size:]
		case 12:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 request snapshot cookies 失败")
			}
			if out.Cookies == nil {
				out.Cookies = make(map[string]string)
			}
			if err := unmarshalStringEntry(value, out.Cookies); err != nil {
				return err
			}
			content = content[size:]
		case 13:
			value, size := protowire.ConsumeBytes(content)
			if size < 0 {
				return gerror.New("解析 request snapshot body 失败")
			}
			out.Body = append([]byte(nil), value...)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("跳过未知 request snapshot 字段失败")
			}
			content = content[size:]
		}
	}
	return nil
}

func marshalIdentitySnapshot(in *IdentitySnapshotV1) []byte {
	var content []byte
	if value := strings.TrimSpace(in.TokenID); value != "" {
		content = appendStringField(content, 1, value)
	}
	if in.UserID != 0 {
		content = appendVarintField(content, 2, uint64(in.UserID))
	}
	if value := strings.TrimSpace(in.Username); value != "" {
		content = appendStringField(content, 3, value)
	}
	if in.Status != 0 {
		content = appendVarintField(content, 4, uint64(in.Status))
	}
	for _, permission := range in.Permissions {
		if normalized := strings.TrimSpace(permission); normalized != "" {
			content = appendStringField(content, 5, normalized)
		}
	}
	for _, roleName := range in.RoleNames {
		if normalized := strings.TrimSpace(roleName); normalized != "" {
			content = appendStringField(content, 6, normalized)
		}
	}
	if in.IsSuperAdmin {
		content = appendVarintField(content, 7, 1)
	}
	return content
}

func unmarshalIdentitySnapshot(content []byte, out *IdentitySnapshotV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("解析 identity snapshot tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 identity snapshot tokenId 失败")
			}
			out.TokenID = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("解析 identity snapshot userId 失败")
			}
			out.UserID = int32(value)
			content = content[size:]
		case 3:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 identity snapshot username 失败")
			}
			out.Username = value
			content = content[size:]
		case 4:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("解析 identity snapshot status 失败")
			}
			out.Status = int32(value)
			content = content[size:]
		case 5:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 identity snapshot permissions 失败")
			}
			out.Permissions = append(out.Permissions, value)
			content = content[size:]
		case 6:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 identity snapshot roleNames 失败")
			}
			out.RoleNames = append(out.RoleNames, value)
			content = content[size:]
		case 7:
			value, size := protowire.ConsumeVarint(content)
			if size < 0 {
				return gerror.New("解析 identity snapshot isSuperAdmin 失败")
			}
			out.IsSuperAdmin = value > 0
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("跳过未知 identity snapshot 字段失败")
			}
			content = content[size:]
		}
	}
	return nil
}

func marshalFailure(in *BridgeFailureV1) []byte {
	var content []byte
	if value := strings.TrimSpace(in.Code); value != "" {
		content = appendStringField(content, 1, value)
	}
	if value := strings.TrimSpace(in.Message); value != "" {
		content = appendStringField(content, 2, value)
	}
	return content
}

func unmarshalFailure(content []byte, out *BridgeFailureV1) error {
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("解析 failure tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 failure code 失败")
			}
			out.Code = value
			content = content[size:]
		case 2:
			value, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 failure message 失败")
			}
			out.Message = value
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("跳过未知 failure 字段失败")
			}
			content = content[size:]
		}
	}
	return nil
}

func appendHeaderMap(content []byte, fieldNumber protowire.Number, values map[string][]string) []byte {
	keys := sortedKeys(values)
	for _, key := range keys {
		entry := marshalStringListPair(key, values[key])
		content = appendBytesField(content, fieldNumber, entry)
	}
	return content
}

func appendStringMap(content []byte, fieldNumber protowire.Number, values map[string]string) []byte {
	keys := sortedKeys(values)
	for _, key := range keys {
		entry := marshalStringPair(key, values[key])
		content = appendBytesField(content, fieldNumber, entry)
	}
	return content
}

func appendStringListMap(content []byte, fieldNumber protowire.Number, values map[string][]string) []byte {
	keys := sortedKeys(values)
	for _, key := range keys {
		entry := marshalStringListPair(key, values[key])
		content = appendBytesField(content, fieldNumber, entry)
	}
	return content
}

func marshalStringPair(key string, value string) []byte {
	var content []byte
	content = appendStringField(content, 1, strings.TrimSpace(key))
	content = appendStringField(content, 2, strings.TrimSpace(value))
	return content
}

func marshalStringListPair(key string, values []string) []byte {
	var content []byte
	content = appendStringField(content, 1, strings.TrimSpace(key))
	for _, value := range values {
		content = appendStringField(content, 2, strings.TrimSpace(value))
	}
	return content
}

func appendStringField(content []byte, fieldNumber protowire.Number, value string) []byte {
	return appendStringFieldContent(content, fieldNumber, value)
}

func appendStringFieldContent(content []byte, fieldNumber protowire.Number, value string) []byte {
	content = protowire.AppendTag(content, fieldNumber, protowire.BytesType)
	return protowire.AppendString(content, value)
}

func appendBytesField(content []byte, fieldNumber protowire.Number, value []byte) []byte {
	content = protowire.AppendTag(content, fieldNumber, protowire.BytesType)
	return protowire.AppendBytes(content, value)
}

func appendVarintField(content []byte, fieldNumber protowire.Number, value uint64) []byte {
	content = protowire.AppendTag(content, fieldNumber, protowire.VarintType)
	return protowire.AppendVarint(content, value)
}

func unmarshalStringEntry(content []byte, output map[string]string) error {
	var (
		key   string
		value string
	)
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("解析 string map entry tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			item, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 string map entry key 失败")
			}
			key = item
			content = content[size:]
		case 2:
			item, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 string map entry value 失败")
			}
			value = item
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("跳过未知 string map entry 字段失败")
			}
			content = content[size:]
		}
	}
	if strings.TrimSpace(key) != "" {
		output[key] = value
	}
	return nil
}

func unmarshalStringListEntry(content []byte, output map[string][]string) error {
	var (
		key    string
		values []string
	)
	for len(content) > 0 {
		fieldNumber, wireType, length := protowire.ConsumeTag(content)
		if length < 0 {
			return gerror.New("解析 string list entry tag 失败")
		}
		content = content[length:]
		switch fieldNumber {
		case 1:
			item, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 string list entry key 失败")
			}
			key = item
			content = content[size:]
		case 2:
			item, size := protowire.ConsumeString(content)
			if size < 0 {
				return gerror.New("解析 string list entry value 失败")
			}
			values = append(values, item)
			content = content[size:]
		default:
			size := protowire.ConsumeFieldValue(fieldNumber, wireType, content)
			if size < 0 {
				return gerror.New("跳过未知 string list entry 字段失败")
			}
			content = content[size:]
		}
	}
	if strings.TrimSpace(key) != "" {
		output[key] = append([]string(nil), values...)
	}
	return nil
}

func unmarshalHeaderEntry(content []byte, output map[string][]string) error {
	return unmarshalStringListEntry(content, output)
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// EncodeBodyBase64 returns a review-friendly body preview for tests and logs.
func EncodeBodyBase64(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(body)
}
