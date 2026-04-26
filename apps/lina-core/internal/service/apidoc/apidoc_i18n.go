// This file localizes generated OpenAPI documents for request-scoped language
// contexts without mutating generated API, DAO, or entity source files.

package apidoc

import (
	"context"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/net/goai"
)

// localizeDocument applies request-locale translations to all user-visible
// OpenAPI metadata after host, source-plugin, and dynamic-plugin routes finish
// projecting into one document.
func (s *serviceImpl) localizeDocument(ctx context.Context, document *goai.OpenApiV3) {
	if document == nil || !s.shouldLocalizeOpenAPI(ctx) {
		return
	}

	localizer := s.newOpenAPILocalizer(ctx, document)
	seenSchemas := make(map[*goai.Schema]struct{})
	document.Info.Title = localizer.translate("core.openapi.info.title", document.Info.Title)
	document.Info.Description = localizer.translate("core.openapi.info.description", document.Info.Description)
	s.localizeExternalDocs(localizer, "core.openapi.externalDocs", document.ExternalDocs)
	s.localizeServers(localizer, "core.openapi.servers", document.Servers)
	s.localizeTags(localizer, "core.openapi.tags", document.Tags)
	s.localizePaths(localizer, document.Paths, seenSchemas)
	s.localizeComponents(localizer, &document.Components, seenSchemas)
}

// shouldLocalizeOpenAPI reports whether this service can apply request-locale
// projections to generated OpenAPI metadata.
func (s *serviceImpl) shouldLocalizeOpenAPI(ctx context.Context) bool {
	if s == nil || s.i18nSvc == nil {
		return false
	}
	if s.bizCtxSvc == nil {
		return false
	}
	bizCtx := s.bizCtxSvc.Get(ctx)
	return bizCtx != nil && strings.TrimSpace(bizCtx.Locale) != ""
}

// openAPILocalizer carries a request-locale catalog and structural lookup
// indexes for stable apidoc i18n keys.
type openAPILocalizer struct {
	locale              string
	catalog             map[string]string
	requestKeyByDesc    map[string]string
	schemaKeyBySchema   map[*goai.Schema]string
	ambiguousRequestDes map[string]struct{}
}

// newOpenAPILocalizer creates the structured-key lookup context used for one
// OpenAPI document projection.
func (s *serviceImpl) newOpenAPILocalizer(ctx context.Context, document *goai.OpenApiV3) *openAPILocalizer {
	localizer := &openAPILocalizer{
		locale:              s.i18nSvc.GetLocale(ctx),
		catalog:             s.loadOpenAPIMessageCatalog(ctx, s.i18nSvc.GetLocale(ctx)),
		requestKeyByDesc:    make(map[string]string),
		schemaKeyBySchema:   make(map[*goai.Schema]string),
		ambiguousRequestDes: make(map[string]struct{}),
	}
	if document == nil {
		return localizer
	}
	for name, ref := range document.Components.Schemas.Map() {
		componentKey := normalizeOpenAPIComponentKey(name)
		if ref.Value == nil || componentKey == "" {
			continue
		}
		localizer.schemaKeyBySchema[ref.Value] = componentKey
		if !strings.HasSuffix(name, "Req") || strings.TrimSpace(ref.Value.Description) == "" {
			continue
		}
		description := ref.Value.Description
		if existingKey, ok := localizer.requestKeyByDesc[description]; ok && existingKey != componentKey {
			localizer.ambiguousRequestDes[description] = struct{}{}
			delete(localizer.requestKeyByDesc, description)
			continue
		}
		if _, ambiguous := localizer.ambiguousRequestDes[description]; !ambiguous {
			localizer.requestKeyByDesc[description] = componentKey
		}
	}
	return localizer
}

// translate resolves one OpenAPI display string by stable structural apidoc key.
func (l *openAPILocalizer) translate(key string, source string) string {
	if l == nil || strings.TrimSpace(source) == "" {
		return source
	}
	if strings.TrimSpace(key) != "" {
		if translated, ok := l.catalog[key]; ok {
			return translated
		}
		if translated, ok := l.catalog[openAPIPluginKeyAlias(key)]; ok {
			return translated
		}
		for _, fallbackKey := range openAPICommonFallbackKeys(key) {
			if translated, ok := l.catalog[fallbackKey]; ok {
				return translated
			}
		}
	}
	return source
}

// openAPICommonFallbackKeys returns shared apidoc translation keys for highly
// repeated generated metadata such as standard response wrappers and paging
// fields. Exact structural keys still win when present.
func openAPICommonFallbackKeys(key string) []string {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return nil
	}

	switch {
	case matchesOpenAPIStandardResponseField(trimmedKey, "code"):
		return []string{"core.common.responses.fields.code.dc"}
	case matchesOpenAPIStandardResponseField(trimmedKey, "message"):
		return []string{"core.common.responses.fields.message.dc"}
	case matchesOpenAPIStandardResponseField(trimmedKey, "data"):
		return []string{"core.common.responses.fields.data.dc"}
	case strings.HasSuffix(trimmedKey, "Res.schema.dc"):
		return []string{"core.common.schemas.response.dc"}
	case strings.HasSuffix(trimmedKey, ".fields.pageNum.dc"):
		return []string{"core.common.fields.pageNum.dc"}
	case strings.HasSuffix(trimmedKey, ".fields.pageSize.dc"):
		return []string{"core.common.fields.pageSize.dc"}
	case strings.HasSuffix(trimmedKey, ".fields.total.dc"):
		return []string{"core.common.fields.total.dc"}
	case strings.HasSuffix(trimmedKey, ".fields.createdAt.dc"):
		return []string{"core.common.fields.createdAt.dc"}
	case strings.HasSuffix(trimmedKey, ".fields.updatedAt.dc"):
		return []string{"core.common.fields.updatedAt.dc"}
	case strings.HasSuffix(trimmedKey, ".fields.deletedAt.dc"):
		return []string{"core.common.fields.deletedAt.dc"}
	default:
		return nil
	}
}

// matchesOpenAPIStandardResponseField reports whether key targets the standard
// response wrapper field emitted under one operation response.
func matchesOpenAPIStandardResponseField(key string, field string) bool {
	segments := strings.Split(strings.TrimSpace(key), ".")
	for index := 0; index+6 < len(segments); index++ {
		if segments[index] != "responses" {
			continue
		}
		if segments[index+2] != "content" || segments[index+4] != "fields" {
			continue
		}
		return segments[index+5] == field && segments[index+6] == "dc" && index+7 == len(segments)
	}
	return false
}

// openAPIPluginKeyAlias supports source-plugin module names such as
// `lina-plugin-demo-source`, where the plugin ID itself starts with `plugin-`.
func openAPIPluginKeyAlias(key string) string {
	parts := strings.Split(strings.TrimSpace(key), ".")
	if len(parts) < 3 || parts[0] != "plugins" || strings.HasPrefix(parts[1], "plugin_") {
		return ""
	}
	parts[1] = "plugin_" + parts[1]
	return strings.Join(parts, ".")
}

// operationBaseKey returns the best stable key base for one operation. Static
// routes prefer their request DTO schema; dynamic routes fall back to path or
// operation-id derived keys.
func (l *openAPILocalizer) operationBaseKey(pathName string, method string, operation *goai.Operation) string {
	if operation == nil {
		return buildOpenAPIPathOperationKey(pathName, method)
	}
	if key := l.requestBodyComponentKey(operation.RequestBody); key != "" {
		return key
	}
	if key := l.requestKeyByDesc[operation.Description]; key != "" {
		return key
	}
	if isDynamicPluginOpenAPIPath(pathName) {
		return buildOpenAPIPathOperationKey(pathName, method)
	}
	if strings.TrimSpace(operation.OperationID) != "" {
		return "core.operations." + sanitizeOpenAPIKeyPart(operation.OperationID)
	}
	return buildOpenAPIPathOperationKey(pathName, method)
}

// requestBodyComponentKey returns the referenced request component key when one
// is available for an operation request body.
func (l *openAPILocalizer) requestBodyComponentKey(ref *goai.RequestBodyRef) string {
	if l == nil || ref == nil || ref.Value == nil {
		return ""
	}
	for _, mediaType := range ref.Value.Content {
		if key := l.schemaRefComponentKey(mediaType.Schema); key != "" {
			return key
		}
	}
	return ""
}

// schemaRefComponentKey resolves a schema ref or schema pointer to its component
// key base.
func (l *openAPILocalizer) schemaRefComponentKey(ref *goai.SchemaRef) string {
	if l == nil || ref == nil {
		return ""
	}
	if ref.Ref != "" {
		return normalizeOpenAPIComponentKey(strings.TrimPrefix(ref.Ref, "#/components/schemas/"))
	}
	if ref.Value != nil {
		return l.schemaKeyBySchema[ref.Value]
	}
	return ""
}

// localizeTags localizes the top-level OpenAPI tag collection when present.
func (s *serviceImpl) localizeTags(localizer *openAPILocalizer, keyBase string, tags *goai.Tags) {
	if tags == nil {
		return
	}
	for index := range *tags {
		itemKey := keyBase + "." + sanitizeOpenAPIKeyPart((*tags)[index].Name)
		(*tags)[index].Name = localizer.translate(itemKey+".name", (*tags)[index].Name)
		(*tags)[index].Description = localizer.translate(itemKey+".dc", (*tags)[index].Description)
		s.localizeExternalDocs(localizer, itemKey+".externalDocs", (*tags)[index].ExternalDocs)
	}
}

// localizePaths localizes every path item and operation in the OpenAPI paths
// table.
func (s *serviceImpl) localizePaths(localizer *openAPILocalizer, paths goai.Paths, seenSchemas map[*goai.Schema]struct{}) {
	for pathName, pathItem := range paths {
		pathKey := buildOpenAPIPathKey(pathName)
		pathItem.Summary = localizer.translate(pathKey+".summary", pathItem.Summary)
		pathItem.Description = localizer.translate(pathKey+".dc", pathItem.Description)
		s.localizeExtensions(localizer, pathKey+".extensions", pathItem.XExtensions)
		s.localizeServers(localizer, pathKey+".servers", &pathItem.Servers)
		s.localizeParameters(localizer, pathKey, pathItem.Parameters, seenSchemas)
		s.localizeOperation(localizer, pathName, "connect", pathItem.Connect, seenSchemas)
		s.localizeOperation(localizer, pathName, "delete", pathItem.Delete, seenSchemas)
		s.localizeOperation(localizer, pathName, "get", pathItem.Get, seenSchemas)
		s.localizeOperation(localizer, pathName, "head", pathItem.Head, seenSchemas)
		s.localizeOperation(localizer, pathName, "options", pathItem.Options, seenSchemas)
		s.localizeOperation(localizer, pathName, "patch", pathItem.Patch, seenSchemas)
		s.localizeOperation(localizer, pathName, "post", pathItem.Post, seenSchemas)
		s.localizeOperation(localizer, pathName, "put", pathItem.Put, seenSchemas)
		s.localizeOperation(localizer, pathName, "trace", pathItem.Trace, seenSchemas)
		paths[pathName] = pathItem
	}
}

// localizeOperation localizes one operation's grouping, descriptions, input
// parameters, request body, responses, and extension metadata.
func (s *serviceImpl) localizeOperation(
	localizer *openAPILocalizer,
	pathName string,
	method string,
	operation *goai.Operation,
	seenSchemas map[*goai.Schema]struct{},
) {
	if operation == nil {
		return
	}
	operationKey := localizer.operationBaseKey(pathName, method, operation)
	for index, tag := range operation.Tags {
		tagKey := operationKey + ".meta.tags"
		if index > 0 {
			tagKey = tagKey + "." + strconv.Itoa(index)
		}
		operation.Tags[index] = localizer.translate(tagKey, tag)
	}
	operation.Summary = localizer.translate(operationKey+".meta.summary", operation.Summary)
	operation.Description = localizer.translate(operationKey+".meta.dc", operation.Description)
	s.localizeParameters(localizer, operationKey, operation.Parameters, seenSchemas)
	s.localizeRequestBodyRef(localizer, operationKey+".requestBody", operation.RequestBody, seenSchemas)
	s.localizeResponses(localizer, operationKey, operation.Responses, seenSchemas)
	s.localizeServers(localizer, operationKey+".servers", operation.Servers)
	s.localizeExternalDocs(localizer, operationKey+".externalDocs", operation.ExternalDocs)
	s.localizeExtensions(localizer, operationKey+".extensions", operation.XExtensions)
}

// localizeComponents localizes reusable schemas, parameters, request bodies,
// responses, security descriptions, examples, headers, and links.
func (s *serviceImpl) localizeComponents(localizer *openAPILocalizer, components *goai.Components, seenSchemas map[*goai.Schema]struct{}) {
	if components == nil {
		return
	}
	s.localizeSchemas(localizer, &components.Schemas, seenSchemas, "core.components.schemas")
	for name, parameter := range components.Parameters {
		s.localizeParameterRef(localizer, "core.components.parameters."+sanitizeOpenAPIKeyPart(name), parameter, seenSchemas)
		components.Parameters[name] = parameter
	}
	for name, requestBody := range components.RequestBodies {
		s.localizeRequestBodyRef(localizer, "core.components.requestBodies."+sanitizeOpenAPIKeyPart(name), requestBody, seenSchemas)
		components.RequestBodies[name] = requestBody
	}
	s.localizeResponses(localizer, "core.components", components.Responses, seenSchemas)
	for name, scheme := range components.SecuritySchemes {
		if scheme.Value != nil {
			schemeKey := "core.openapi.securitySchemes." + sanitizeOpenAPIKeyPart(name)
			scheme.Value.Description = localizer.translate(schemeKey+".dc", scheme.Value.Description)
			s.localizeOAuthFlows(localizer, schemeKey+".flows", scheme.Value.Flows)
		}
		components.SecuritySchemes[name] = scheme
	}
	s.localizeExamples(localizer, "core.components.examples", components.Examples)
	s.localizeHeaders(localizer, "core.components.headers", components.Headers, seenSchemas)
	s.localizeLinks(localizer, "core.components.links", components.Links)
}

// localizeSchemas localizes schema refs stored in GoFrame's ordered schema map.
func (s *serviceImpl) localizeSchemas(
	localizer *openAPILocalizer,
	schemas *goai.Schemas,
	seenSchemas map[*goai.Schema]struct{},
	keyBase string,
) {
	if schemas == nil {
		return
	}
	for name, ref := range schemas.Map() {
		schemaKey := normalizeOpenAPIComponentKey(name)
		if schemaKey == "" {
			schemaKey = keyBase + "." + sanitizeOpenAPIKeyPart(name)
		}
		s.localizeSchemaRef(localizer, &ref, schemaKey+".schema", seenSchemas)
		schemas.Set(name, ref)
	}
}

// localizeSchemaRef localizes one schema ref description and its inline schema
// value when present.
func (s *serviceImpl) localizeSchemaRef(
	localizer *openAPILocalizer,
	ref *goai.SchemaRef,
	keyBase string,
	seenSchemas map[*goai.Schema]struct{},
) {
	if ref == nil {
		return
	}
	if componentKey := localizer.schemaRefComponentKey(ref); componentKey != "" && strings.HasSuffix(keyBase, ".schema") {
		keyBase = componentKey + ".schema"
	}
	ref.Description = localizer.translate(keyBase+".dc", ref.Description)
	s.localizeSchema(localizer, ref.Value, keyBase, seenSchemas)
}

// localizeSchema recursively localizes all user-visible schema metadata while
// avoiding infinite loops on shared schema pointers.
func (s *serviceImpl) localizeSchema(
	localizer *openAPILocalizer,
	schema *goai.Schema,
	keyBase string,
	seenSchemas map[*goai.Schema]struct{},
) {
	if schema == nil {
		return
	}

	if componentKey := localizer.schemaKeyBySchema[schema]; componentKey != "" && strings.HasSuffix(keyBase, ".schema") {
		keyBase = componentKey + ".schema"
	}

	schema.Title = localizer.translate(keyBase+".title", schema.Title)
	schema.Description = localizer.translate(keyBase+".dc", schema.Description)
	schema.Default = s.localizeOpenAPIValue(localizer, keyBase+".default", schema.Default)
	s.localizeExternalDocs(localizer, keyBase+".externalDocs", schema.ExternalDocs)
	s.localizeExtensions(localizer, keyBase+".extensions", schema.XExtensions)

	if _, ok := seenSchemas[schema]; ok {
		return
	}
	seenSchemas[schema] = struct{}{}

	for index := range schema.OneOf {
		s.localizeSchemaRef(localizer, &schema.OneOf[index], keyBase+".oneOf."+strconv.Itoa(index), seenSchemas)
	}
	for index := range schema.AnyOf {
		s.localizeSchemaRef(localizer, &schema.AnyOf[index], keyBase+".anyOf."+strconv.Itoa(index), seenSchemas)
	}
	for index := range schema.AllOf {
		s.localizeSchemaRef(localizer, &schema.AllOf[index], keyBase+".allOf."+strconv.Itoa(index), seenSchemas)
	}
	s.localizeSchemaRef(localizer, schema.Not, keyBase+".not", seenSchemas)
	s.localizeSchemaRef(localizer, schema.Items, keyBase+".items", seenSchemas)
	s.localizeSchemaRef(localizer, schema.AdditionalProperties, keyBase+".additionalProperties", seenSchemas)
	if schema.Properties != nil {
		for propertyName, propertyRef := range schema.Properties.Map() {
			propertyKey := buildOpenAPIFieldKey(keyBase, propertyName)
			s.localizeSchemaRef(localizer, &propertyRef, propertyKey, seenSchemas)
			schema.Properties.Set(propertyName, propertyRef)
		}
	}
}

// localizeParameters localizes a parameter list in place.
func (s *serviceImpl) localizeParameters(
	localizer *openAPILocalizer,
	operationKey string,
	parameters goai.Parameters,
	seenSchemas map[*goai.Schema]struct{},
) {
	for index := range parameters {
		parameterKey := operationKey + ".parameters." + strconv.Itoa(index)
		if parameters[index].Value != nil && parameters[index].Value.Name != "" {
			parameterKey = operationKey + ".fields." + sanitizeOpenAPIKeyPart(parameters[index].Value.Name)
		}
		s.localizeParameterRef(localizer, parameterKey, &parameters[index], seenSchemas)
	}
}

// localizeParameterRef localizes one parameter ref and its schema/content
// details.
func (s *serviceImpl) localizeParameterRef(
	localizer *openAPILocalizer,
	keyBase string,
	ref *goai.ParameterRef,
	seenSchemas map[*goai.Schema]struct{},
) {
	if ref == nil || ref.Value == nil {
		return
	}
	ref.Value.Description = localizer.translate(keyBase+".dc", ref.Value.Description)
	s.localizeSchemaRef(localizer, ref.Value.Schema, keyBase, seenSchemas)
	if ref.Value.Content != nil {
		s.localizeContent(localizer, keyBase+".content", *ref.Value.Content, seenSchemas)
	}
	if ref.Value.Examples != nil {
		s.localizeExamples(localizer, keyBase+".examples", *ref.Value.Examples)
	}
	s.localizeExtensions(localizer, keyBase+".extensions", ref.Value.XExtensions)
}

// localizeRequestBodyRef localizes request body metadata and JSON schema
// content.
func (s *serviceImpl) localizeRequestBodyRef(
	localizer *openAPILocalizer,
	keyBase string,
	ref *goai.RequestBodyRef,
	seenSchemas map[*goai.Schema]struct{},
) {
	if ref == nil || ref.Value == nil {
		return
	}
	ref.Value.Description = localizer.translate(keyBase+".dc", ref.Value.Description)
	s.localizeContent(localizer, keyBase+".content", ref.Value.Content, seenSchemas)
}

// localizeResponses localizes every response in one responses map.
func (s *serviceImpl) localizeResponses(
	localizer *openAPILocalizer,
	keyBase string,
	responses goai.Responses,
	seenSchemas map[*goai.Schema]struct{},
) {
	for code, response := range responses {
		s.localizeResponseRef(localizer, keyBase+".responses."+sanitizeOpenAPIKeyPart(code), &response, seenSchemas)
		responses[code] = response
	}
}

// localizeResponseRef localizes one response ref, including headers, content,
// links, and extension metadata.
func (s *serviceImpl) localizeResponseRef(
	localizer *openAPILocalizer,
	keyBase string,
	ref *goai.ResponseRef,
	seenSchemas map[*goai.Schema]struct{},
) {
	if ref == nil || ref.Value == nil {
		return
	}
	ref.Value.Description = localizer.translate(keyBase+".dc", ref.Value.Description)
	s.localizeHeaders(localizer, keyBase+".headers", ref.Value.Headers, seenSchemas)
	s.localizeContent(localizer, keyBase+".content", ref.Value.Content, seenSchemas)
	s.localizeLinks(localizer, keyBase+".links", ref.Value.Links)
	s.localizeExtensions(localizer, keyBase+".extensions", ref.Value.XExtensions)
}

// localizeContent localizes schema and example metadata for media-type content.
func (s *serviceImpl) localizeContent(
	localizer *openAPILocalizer,
	keyBase string,
	content goai.Content,
	seenSchemas map[*goai.Schema]struct{},
) {
	for contentType, mediaType := range content {
		contentKey := keyBase + "." + sanitizeOpenAPIKeyPart(contentType)
		s.localizeSchemaRef(localizer, mediaType.Schema, contentKey+".schema", seenSchemas)
		s.localizeExamples(localizer, contentKey+".examples", mediaType.Examples)
		content[contentType] = mediaType
	}
}

// localizeExamples localizes example summaries and descriptions while leaving
// example values unchanged.
func (s *serviceImpl) localizeExamples(localizer *openAPILocalizer, keyBase string, examples goai.Examples) {
	for name, ref := range examples {
		if ref == nil || ref.Value == nil {
			continue
		}
		exampleKey := keyBase + "." + sanitizeOpenAPIKeyPart(name)
		ref.Value.Summary = localizer.translate(exampleKey+".summary", ref.Value.Summary)
		ref.Value.Description = localizer.translate(exampleKey+".dc", ref.Value.Description)
		examples[name] = ref
	}
}

// localizeHeaders localizes reusable or response-specific header metadata.
func (s *serviceImpl) localizeHeaders(
	localizer *openAPILocalizer,
	keyBase string,
	headers goai.Headers,
	seenSchemas map[*goai.Schema]struct{},
) {
	for name, header := range headers {
		if header.Value != nil {
			s.localizeParameter(localizer, keyBase+"."+sanitizeOpenAPIKeyPart(name), &header.Value.Parameter, seenSchemas)
		}
		headers[name] = header
	}
}

// localizeParameter localizes one concrete parameter object.
func (s *serviceImpl) localizeParameter(
	localizer *openAPILocalizer,
	keyBase string,
	parameter *goai.Parameter,
	seenSchemas map[*goai.Schema]struct{},
) {
	if parameter == nil {
		return
	}
	parameter.Description = localizer.translate(keyBase+".dc", parameter.Description)
	s.localizeSchemaRef(localizer, parameter.Schema, keyBase, seenSchemas)
	if parameter.Content != nil {
		s.localizeContent(localizer, keyBase+".content", *parameter.Content, seenSchemas)
	}
	if parameter.Examples != nil {
		s.localizeExamples(localizer, keyBase+".examples", *parameter.Examples)
	}
	s.localizeExtensions(localizer, keyBase+".extensions", parameter.XExtensions)
}

// localizeLinks localizes link descriptions and nested server metadata.
func (s *serviceImpl) localizeLinks(localizer *openAPILocalizer, keyBase string, links goai.Links) {
	for name, link := range links {
		if link.Value != nil {
			linkKey := keyBase + "." + sanitizeOpenAPIKeyPart(name)
			link.Value.Description = localizer.translate(linkKey+".dc", link.Value.Description)
			link.Value.RequestBody = s.localizeOpenAPIValue(localizer, linkKey+".requestBody", link.Value.RequestBody)
			if link.Value.Server != nil {
				servers := goai.Servers{*link.Value.Server}
				s.localizeServers(localizer, linkKey+".server", &servers)
				link.Value.Server = &servers[0]
			}
		}
		links[name] = link
	}
}

// localizeServers localizes server and server-variable descriptions.
func (s *serviceImpl) localizeServers(localizer *openAPILocalizer, keyBase string, servers *goai.Servers) {
	if servers == nil {
		return
	}
	for index := range *servers {
		serverKey := keyBase + "." + strconv.Itoa(index)
		(*servers)[index].Description = localizer.translate(serverKey+".dc", (*servers)[index].Description)
		for name, variable := range (*servers)[index].Variables {
			if variable == nil {
				continue
			}
			variable.Description = localizer.translate(serverKey+".variables."+sanitizeOpenAPIKeyPart(name)+".dc", variable.Description)
			(*servers)[index].Variables[name] = variable
		}
	}
}

// localizeExternalDocs localizes optional external-document descriptions.
func (s *serviceImpl) localizeExternalDocs(localizer *openAPILocalizer, keyBase string, docs *goai.ExternalDocs) {
	if docs == nil {
		return
	}
	docs.Description = localizer.translate(keyBase+".dc", docs.Description)
}

// localizeOAuthFlows localizes OAuth scope descriptions when a security scheme
// uses them.
func (s *serviceImpl) localizeOAuthFlows(localizer *openAPILocalizer, keyBase string, flows *goai.OAuthFlows) {
	if flows == nil {
		return
	}
	s.localizeOAuthFlow(localizer, keyBase+".implicit", flows.Implicit)
	s.localizeOAuthFlow(localizer, keyBase+".password", flows.Password)
	s.localizeOAuthFlow(localizer, keyBase+".clientCredentials", flows.ClientCredentials)
	s.localizeOAuthFlow(localizer, keyBase+".authorizationCode", flows.AuthorizationCode)
}

// localizeOAuthFlow localizes one OAuth flow's scope descriptions.
func (s *serviceImpl) localizeOAuthFlow(localizer *openAPILocalizer, keyBase string, flow *goai.OAuthFlow) {
	if flow == nil {
		return
	}
	for scope, description := range flow.Scopes {
		flow.Scopes[scope] = localizer.translate(keyBase+".scopes."+sanitizeOpenAPIKeyPart(scope)+".dc", description)
	}
}

// localizeExtensions localizes string-valued x-extension metadata emitted into
// the OpenAPI JSON.
func (s *serviceImpl) localizeExtensions(localizer *openAPILocalizer, keyBase string, extensions goai.XExtensions) {
	for key, value := range extensions {
		extensions[key] = localizer.translate(keyBase+"."+sanitizeOpenAPIKeyPart(key), value)
	}
}

// localizeOpenAPIValue localizes OpenAPI metadata values such as defaults while
// preserving non-string value types and example values.
func (s *serviceImpl) localizeOpenAPIValue(localizer *openAPILocalizer, keyBase string, value any) any {
	switch typedValue := value.(type) {
	case string:
		return localizer.translate(keyBase, typedValue)
	case []any:
		for index, item := range typedValue {
			typedValue[index] = s.localizeOpenAPIValue(localizer, keyBase+"."+strconv.Itoa(index), item)
		}
		return typedValue
	case map[string]any:
		for key, item := range typedValue {
			typedValue[key] = s.localizeOpenAPIValue(localizer, keyBase+"."+sanitizeOpenAPIKeyPart(key), item)
		}
		return typedValue
	default:
		return value
	}
}

// normalizeOpenAPIComponentKey converts GoFrame schema component names into
// stable, repository-readable apidoc keys.
func normalizeOpenAPIComponentKey(name string) string {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return ""
	}
	if strings.HasPrefix(trimmedName, "lina-plugin-") {
		return normalizeSourcePluginOpenAPIComponentKey(trimmedName)
	}
	replacements := []struct {
		old string
		new string
	}{
		{old: "lina-core.", new: "core."},
		{old: "lina-plugins.", new: "plugins."},
	}
	for _, replacement := range replacements {
		if strings.HasPrefix(trimmedName, replacement.old) {
			trimmedName = replacement.new + strings.TrimPrefix(trimmedName, replacement.old)
			break
		}
	}
	return sanitizeOpenAPIKey(trimmedName)
}

// normalizeSourcePluginOpenAPIComponentKey converts source-plugin schema
// component names into plugin-owned structural apidoc keys.
func normalizeSourcePluginOpenAPIComponentKey(name string) string {
	trimmedName := strings.TrimSpace(name)
	withoutPrefix := strings.TrimPrefix(trimmedName, "lina-plugin-")
	pluginPart, rest, ok := strings.Cut(withoutPrefix, ".")
	if !ok || strings.TrimSpace(pluginPart) == "" {
		return sanitizeOpenAPIKey(trimmedName)
	}
	if strings.HasPrefix(rest, "backend.api.") {
		rest = strings.TrimPrefix(rest, "backend.")
	}
	return "plugins." + sanitizeOpenAPIKeyPart(pluginPart) + "." + sanitizeOpenAPIKey(rest)
}

// buildOpenAPIFieldKey returns the structured key base for one schema property
// or request parameter.
func buildOpenAPIFieldKey(parentKey string, fieldName string) string {
	fieldKey := sanitizeOpenAPIKeyPart(fieldName)
	if strings.HasSuffix(parentKey, ".schema") {
		return strings.TrimSuffix(parentKey, ".schema") + ".fields." + fieldKey
	}
	return parentKey + ".fields." + fieldKey
}

// buildOpenAPIPathKey returns a stable structural key base for a path item.
func buildOpenAPIPathKey(pathName string) string {
	return "core.paths." + sanitizeOpenAPIPathKey(pathName)
}

// buildOpenAPIPathOperationKey returns a stable structural key base for an
// operation when no request DTO component can be inferred.
func buildOpenAPIPathOperationKey(pathName string, method string) string {
	segments := openAPIPathSegments(pathName)
	if isDynamicPluginOpenAPIPath(pathName) {
		pluginID := sanitizeOpenAPIKeyPart(segments[3])
		remainingPath := strings.Join(segments[4:], ".")
		if remainingPath == "" {
			remainingPath = "root"
		}
		return "plugins." + pluginID + ".paths." + sanitizeOpenAPIKeyPart(method) + "." + remainingPath
	}
	return buildOpenAPIPathKey(pathName) + "." + sanitizeOpenAPIKeyPart(method)
}

// isDynamicPluginOpenAPIPath reports whether a public OpenAPI path belongs to
// the dynamic-plugin extension dispatch namespace.
func isDynamicPluginOpenAPIPath(pathName string) bool {
	segments := openAPIPathSegments(pathName)
	return len(segments) >= 4 && slices.Equal(segments[:3], []string{"api", "v1", "extensions"})
}

// sanitizeOpenAPIPathKey converts an OpenAPI path into dot-separated key parts.
func sanitizeOpenAPIPathKey(pathName string) string {
	segments := openAPIPathSegments(pathName)
	if len(segments) == 0 {
		return "root"
	}
	return strings.Join(segments, ".")
}

// openAPIPathSegments normalizes path segments for use in translation keys.
func openAPIPathSegments(pathName string) []string {
	var segments []string
	for _, segment := range strings.Split(strings.Trim(pathName, "/"), "/") {
		if strings.TrimSpace(segment) == "" {
			continue
		}
		segments = append(segments, sanitizeOpenAPIKeyPart(segment))
	}
	return segments
}

// sanitizeOpenAPIKey normalizes a full key while preserving dots as hierarchy
// separators.
func sanitizeOpenAPIKey(key string) string {
	parts := strings.Split(strings.TrimSpace(key), ".")
	for index, part := range parts {
		parts[index] = sanitizeOpenAPIKeyPart(part)
	}
	return strings.Join(parts, ".")
}

var openAPIKeyInvalidCharsPattern = regexp.MustCompile(`[^A-Za-z0-9_]+`)

// sanitizeOpenAPIKeyPart normalizes one key segment for safe JSON-object keys.
func sanitizeOpenAPIKeyPart(part string) string {
	trimmedPart := strings.TrimSpace(part)
	trimmedPart = strings.Trim(trimmedPart, "{}")
	trimmedPart = strings.ReplaceAll(trimmedPart, "-", "_")
	trimmedPart = openAPIKeyInvalidCharsPattern.ReplaceAllString(trimmedPart, "_")
	trimmedPart = strings.Trim(trimmedPart, "_")
	if trimmedPart == "" {
		return "empty"
	}
	return trimmedPart
}
