package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Swagger20 represents Swagger 2.0 specification
type Swagger20 struct {
	Swagger             string                `json:"swagger"`
	Info                Info                  `json:"info"`
	Host                string                `json:"host,omitempty"`
	BasePath            string                `json:"basePath,omitempty"`
	Schemes             []string              `json:"schemes,omitempty"`
	Consumes            []string              `json:"consumes,omitempty"`
	Produces            []string              `json:"produces,omitempty"`
	Paths               map[string]PathItem   `json:"paths"`
	Definitions         map[string]Definition `json:"definitions,omitempty"`
	Parameters          map[string]Parameter  `json:"parameters,omitempty"`
	Responses           map[string]Response   `json:"responses,omitempty"`
	SecurityDefinitions map[string]Security   `json:"securityDefinitions,omitempty"`
	Security            []map[string][]string `json:"security,omitempty"`
	Tags                []Tag                 `json:"tags,omitempty"`
	ExternalDocs        *ExternalDocs         `json:"externalDocs,omitempty"`
}

// OpenAPI31 represents OpenAPI 3.1.0 specification
type OpenAPI31 struct {
	OpenAPI      string                `json:"openapi" yaml:"openapi"`
	Info         Info                  `json:"info" yaml:"info"`
	Servers      []Server              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths        map[string]PathItem31 `json:"paths,omitempty" yaml:"paths,omitempty"`
	Components   *Components           `json:"components,omitempty" yaml:"components,omitempty"`
	Security     []map[string][]string `json:"security,omitempty" yaml:"security,omitempty"`
	Tags         []Tag                 `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocs         `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// Info represents API information
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string   `json:"version" yaml:"version"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License represents license information
type License struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server represents OpenAPI 3.1 server
type Server struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// PathItem represents Swagger 2.0 path item
type PathItem struct {
	Get        *Operation  `json:"get,omitempty"`
	Put        *Operation  `json:"put,omitempty"`
	Post       *Operation  `json:"post,omitempty"`
	Delete     *Operation  `json:"delete,omitempty"`
	Options    *Operation  `json:"options,omitempty"`
	Head       *Operation  `json:"head,omitempty"`
	Patch      *Operation  `json:"patch,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"`
}

// PathItem31 represents OpenAPI 3.1 path item
type PathItem31 struct {
	Get        *Operation31   `json:"get,omitempty" yaml:"get,omitempty"`
	Put        *Operation31   `json:"put,omitempty" yaml:"put,omitempty"`
	Post       *Operation31   `json:"post,omitempty" yaml:"post,omitempty"`
	Delete     *Operation31   `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options    *Operation31   `json:"options,omitempty" yaml:"options,omitempty"`
	Head       *Operation31   `json:"head,omitempty" yaml:"head,omitempty"`
	Patch      *Operation31   `json:"patch,omitempty" yaml:"patch,omitempty"`
	Parameters []Parameter31  `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Summary    string         `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
}

// Operation represents Swagger 2.0 operation
type Operation struct {
	Tags        []string              `json:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
	Consumes    []string              `json:"consumes,omitempty"`
	Produces    []string              `json:"produces,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Deprecated  bool                  `json:"deprecated,omitempty"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Operation31 represents OpenAPI 3.1 operation
type Operation31 struct {
	Tags        []string              `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary     string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	OperationID string                `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters  []Parameter31         `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response31 `json:"responses" yaml:"responses"`
	Deprecated  bool                  `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security    []map[string][]string `json:"security,omitempty" yaml:"security,omitempty"`
}

// Parameter represents Swagger 2.0 parameter
type Parameter struct {
	Name            string        `json:"name"`
	In              string        `json:"in"`
	Description     string        `json:"description,omitempty"`
	Required        bool          `json:"required,omitempty"`
	Schema          *Schema       `json:"schema,omitempty"`
	Type            string        `json:"type,omitempty"`
	Format          string        `json:"format,omitempty"`
	Items           *Items        `json:"items,omitempty"`
	Enum            []interface{} `json:"enum,omitempty"`
	Default         interface{}   `json:"default,omitempty"`
	Example         interface{}   `json:"example,omitempty"`
	Body            *Schema       `json:"body,omitempty"`
	AllowEmptyValue bool          `json:"allowEmptyValue,omitempty"`
	UniqueItems     bool          `json:"uniqueItems,omitempty"`
}

// Parameter31 represents OpenAPI 3.1 parameter
type Parameter31 struct {
	Name        string      `json:"name" yaml:"name"`
	In          string      `json:"in" yaml:"in"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool        `json:"required,omitempty" yaml:"required,omitempty"`
	Schema      *Schema     `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example     interface{} `json:"example,omitempty" yaml:"example,omitempty"`
}

// RequestBody represents OpenAPI 3.1 request body
type RequestBody struct {
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
	Content     map[string]MediaType `json:"content" yaml:"content"`
}

// MediaType represents OpenAPI 3.1 media type
type MediaType struct {
	Schema  *Schema     `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example interface{} `json:"example,omitempty" yaml:"example,omitempty"`
}

// Response represents Swagger 2.0 response
type Response struct {
	Description string                 `json:"description"`
	Schema      *Schema                `json:"schema,omitempty"`
	Headers     map[string]Header      `json:"headers,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty"`
}

// Response31 represents OpenAPI 3.1 response
type Response31 struct {
	Description string               `json:"description" yaml:"description"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Headers     map[string]Header    `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// Header represents header definition
type Header struct {
	Description string  `json:"description,omitempty"`
	Type        string  `json:"type,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// Schema represents JSON Schema
type Schema struct {
	Type                 string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Format               string                 `json:"format,omitempty" yaml:"format,omitempty"`
	Description          string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Title                string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Items                *Schema                `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           map[string]*Schema     `json:"properties,omitempty" yaml:"properties,omitempty"`
	Required             []string               `json:"required,omitempty" yaml:"required,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default              interface{}            `json:"default,omitempty" yaml:"default,omitempty"`
	Example              interface{}            `json:"example,omitempty" yaml:"example,omitempty"`
	Ref                  string                 `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	AllOf                []Schema               `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	AnyOf                []Schema               `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	OneOf                []Schema               `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	Not                  *Schema                `json:"not,omitempty" yaml:"not,omitempty"`
	AdditionalProperties *AdditionalProperties  `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Nullable             bool                   `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	ReadOnly             bool                   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool                   `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	MinLength            int                    `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength            int                    `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Minimum              float64                `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum              float64                `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	MinItems             int                    `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	MaxItems             int                    `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	UniqueItems          bool                   `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
}

// AdditionalProperties can be either a boolean or a Schema
type AdditionalProperties struct {
	IsBool   bool
	BoolVal  bool
	SchemaVal *Schema
}

// UnmarshalJSON implements custom unmarshaling for AdditionalProperties
func (ap *AdditionalProperties) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as boolean first
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		ap.IsBool = true
		ap.BoolVal = b
		return nil
	}
	
	// Try to unmarshal as Schema
	var s Schema
	if err := json.Unmarshal(data, &s); err == nil {
		ap.IsBool = false
		ap.SchemaVal = &s
		return nil
	}
	
	return fmt.Errorf("AdditionalProperties must be either boolean or Schema")
}

// MarshalJSON implements custom marshaling for AdditionalProperties
func (ap *AdditionalProperties) MarshalJSON() ([]byte, error) {
	if ap.IsBool {
		return json.Marshal(ap.BoolVal)
	}
	if ap.SchemaVal != nil {
		return json.Marshal(ap.SchemaVal)
	}
	return json.Marshal(nil)
}

// Items represents array items
type Items struct {
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
	Ref    string `json:"$ref,omitempty"`
}

// Definition represents Swagger 2.0 definition
type Definition struct {
	Type                 string                 `json:"type,omitempty"`
	Format               string                 `json:"format,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Title                string                 `json:"title,omitempty"`
	Items                *Schema                `json:"items,omitempty"`
	Properties           map[string]*Schema     `json:"properties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty"`
	Default              interface{}            `json:"default,omitempty"`
	Example              interface{}            `json:"example,omitempty"`
	AllOf                []Schema               `json:"allOf,omitempty"`
	AdditionalProperties *AdditionalProperties  `json:"additionalProperties,omitempty"`
}

// Components represents OpenAPI 3.1 components
type Components struct {
	Schemas         map[string]*Schema   `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Parameters      map[string]Parameter31 `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Responses       map[string]Response31  `json:"responses,omitempty" yaml:"responses,omitempty"`
	SecuritySchemes map[string]Security31  `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
}

// Security represents Swagger 2.0 security scheme
type Security struct {
	Type             string            `json:"type"`
	Description      string            `json:"description,omitempty"`
	Name             string            `json:"name,omitempty"`
	In               string            `json:"in,omitempty"`
	Flow             string            `json:"flow,omitempty"`
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty"`
}

// Security31 represents OpenAPI 3.1 security scheme
type Security31 struct {
	Type         string         `json:"type" yaml:"type"`
	Description  string         `json:"description,omitempty" yaml:"description,omitempty"`
	Name         string         `json:"name,omitempty" yaml:"name,omitempty"`
	In           string         `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme       string         `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat string         `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows        *SecurityFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
}

// SecurityFlows represents OAuth2 flows
type SecurityFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow represents OAuth2 flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes" yaml:"scopes"`
}

// Tag represents tag definition
type Tag struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

// ExternalDocs represents external documentation
type ExternalDocs struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

func main() {
	inputFile := flag.String("input", "", "Input Swagger 2.0 JSON file")
	outputDir := flag.String("output", "", "Output directory for OpenAPI 3.1 files")
	flag.Parse()

	if *inputFile == "" || *outputDir == "" {
		log.Fatal("Usage: openapi_convert -input <swagger.json> -output <output_dir>")
	}

	// Read Swagger 2.0 file
	swaggerData, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading Swagger file: %v", err)
	}

	var swagger Swagger20
	if err := json.Unmarshal(swaggerData, &swagger); err != nil {
		log.Fatalf("Error parsing Swagger JSON: %v", err)
	}

	// Convert to OpenAPI 3.1
	openapi := convertToOpenAPI31(swagger)

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Write JSON output
	jsonPath := filepath.Join(*outputDir, "openapi.json")
	jsonData, err := json.MarshalIndent(openapi, "", "    ")
	if err != nil {
		log.Fatalf("Error marshaling OpenAPI JSON: %v", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		log.Fatalf("Error writing OpenAPI JSON: %v", err)
	}

	// Write YAML output
	yamlPath := filepath.Join(*outputDir, "openapi.yaml")
	yamlData, err := yaml.Marshal(openapi)
	if err != nil {
		log.Fatalf("Error marshaling OpenAPI YAML: %v", err)
	}
	if err := os.WriteFile(yamlPath, yamlData, 0644); err != nil {
		log.Fatalf("Error writing OpenAPI YAML: %v", err)
	}

	fmt.Printf("✅ Successfully converted to OpenAPI 3.1\n")
	fmt.Printf("   JSON: %s\n", jsonPath)
	fmt.Printf("   YAML: %s\n", yamlPath)
}

func convertToOpenAPI31(swagger Swagger20) OpenAPI31 {
	servers := buildServers(swagger)
	paths := convertPaths(swagger.Paths)
	components := convertComponents(swagger)

	var security []map[string][]string
	if len(swagger.Security) > 0 {
		security = swagger.Security
	}

	return OpenAPI31{
		OpenAPI:      "3.1.0",
		Info:         swagger.Info,
		Servers:      servers,
		Paths:        paths,
		Components:   components,
		Security:     security,
		Tags:         swagger.Tags,
		ExternalDocs: swagger.ExternalDocs,
	}
}

func buildServers(swagger Swagger20) []Server {
	schemes := swagger.Schemes
	if len(schemes) == 0 {
		schemes = []string{"http"}
	}

	host := swagger.Host
	if host == "" {
		host = "localhost"
	}

	basePath := swagger.BasePath
	if basePath == "" {
		basePath = "/"
	}

	var servers []Server
	for _, scheme := range schemes {
		url := fmt.Sprintf("%s://%s%s", scheme, host, basePath)
		servers = append(servers, Server{
			URL:         url,
			Description: fmt.Sprintf("%s server", strings.Title(scheme)),
		})
	}

	return servers
}

func convertPaths(paths map[string]PathItem) map[string]PathItem31 {
	result := make(map[string]PathItem31)

	for path, item := range paths {
		path31 := PathItem31{
			Parameters: convertParameters31(item.Parameters),
		}

		if item.Get != nil {
			path31.Get = convertOperation31(item.Get)
		}
		if item.Put != nil {
			path31.Put = convertOperation31(item.Put)
		}
		if item.Post != nil {
			path31.Post = convertOperation31(item.Post)
		}
		if item.Delete != nil {
			path31.Delete = convertOperation31(item.Delete)
		}
		if item.Options != nil {
			path31.Options = convertOperation31(item.Options)
		}
		if item.Head != nil {
			path31.Head = convertOperation31(item.Head)
		}
		if item.Patch != nil {
			path31.Patch = convertOperation31(item.Patch)
		}

		result[path] = path31
	}

	return result
}

func convertOperation31(op *Operation) *Operation31 {
	if op == nil {
		return nil
	}

	var params31 []Parameter31
	var requestBody *RequestBody

	for _, param := range op.Parameters {
		if param.In == "body" {
			requestBody = &RequestBody{
				Description: param.Description,
				Required:    param.Required,
				Content:     make(map[string]MediaType),
			}

			consumes := op.Consumes
			if len(consumes) == 0 {
				consumes = []string{"application/json"}
			}

			schema := param.Schema
			if schema == nil {
				schema = param.Body
			}

			for _, ct := range consumes {
				requestBody.Content[ct] = MediaType{
					Schema: convertSchema(schema),
				}
			}
		} else {
			params31 = append(params31, convertParameter31(param))
		}
	}

	responses31 := convertResponses31(op.Responses)

	return &Operation31{
		Tags:        op.Tags,
		Summary:     op.Summary,
		Description: op.Description,
		OperationID: op.OperationID,
		Parameters:  params31,
		RequestBody: requestBody,
		Responses:   responses31,
		Deprecated:  op.Deprecated,
		Security:    op.Security,
	}
}

func convertParameter31(param Parameter) Parameter31 {
	param31 := Parameter31{
		Name:        param.Name,
		In:          param.In,
		Description: param.Description,
		Required:    param.Required,
		Example:     param.Example,
	}

	if param.Schema != nil {
		param31.Schema = convertSchema(param.Schema)
	} else {
		schema := &Schema{
			Type:    param.Type,
			Format:  param.Format,
			Enum:    param.Enum,
			Default: param.Default,
		}

		if param.Items != nil {
			schema.Items = &Schema{
				Type:   param.Items.Type,
				Format: param.Items.Format,
			}
			if param.Items.Ref != "" {
				schema.Items.Ref = convertRef(param.Items.Ref)
			}
		}

		param31.Schema = schema
	}

	return param31
}

func convertParameters31(params []Parameter) []Parameter31 {
	if len(params) == 0 {
		return nil
	}

	result := make([]Parameter31, 0, len(params))
	for _, param := range params {
		if param.In != "body" {
			result = append(result, convertParameter31(param))
		}
	}

	return result
}

func convertResponses31(responses map[string]Response) map[string]Response31 {
	result := make(map[string]Response31)

	for code, resp := range responses {
		response31 := Response31{
			Description: resp.Description,
			Headers:     resp.Headers,
		}

		if resp.Schema != nil {
			response31.Content = map[string]MediaType{
				"application/json": {
					Schema: convertSchema(resp.Schema),
				},
			}
		}

		result[code] = response31
	}

	return result
}

func convertComponents(swagger Swagger20) *Components {
	components := &Components{}

	if len(swagger.Definitions) > 0 {
		components.Schemas = make(map[string]*Schema)
		for name, def := range swagger.Definitions {
			components.Schemas[name] = convertDefinitionToSchema(def)
		}
	}

	if len(swagger.SecurityDefinitions) > 0 {
		components.SecuritySchemes = make(map[string]Security31)
		for name, sec := range swagger.SecurityDefinitions {
			components.SecuritySchemes[name] = convertSecurityDefinition(sec)
		}
	}

	return components
}

func convertDefinitionToSchema(def Definition) *Schema {
	schema := &Schema{
		Type:        def.Type,
		Format:      def.Format,
		Description: def.Description,
		Title:       def.Title,
		Required:    def.Required,
		Enum:        def.Enum,
		Default:     def.Default,
		Example:     def.Example,
	}

	if def.Items != nil {
		schema.Items = convertSchema(def.Items)
	}

	if len(def.Properties) > 0 {
		schema.Properties = make(map[string]*Schema)
		for name, prop := range def.Properties {
			schema.Properties[name] = convertSchema(prop)
		}
	}

	if len(def.AllOf) > 0 {
		schema.AllOf = make([]Schema, len(def.AllOf))
		for i, allOf := range def.AllOf {
			schema.AllOf[i] = *convertSchema(&allOf)
		}
	}

	if def.AdditionalProperties != nil {
		schema.AdditionalProperties = &AdditionalProperties{
			IsBool:    def.AdditionalProperties.IsBool,
			BoolVal:   def.AdditionalProperties.BoolVal,
			SchemaVal: convertSchema(def.AdditionalProperties.SchemaVal),
		}
	}

	return schema
}

func convertSchema(schema *Schema) *Schema {
	if schema == nil {
		return nil
	}

	result := &Schema{
		Type:                 schema.Type,
		Format:               schema.Format,
		Description:          schema.Description,
		Title:                schema.Title,
		Required:             schema.Required,
		Enum:                 schema.Enum,
		Default:              schema.Default,
		Example:              schema.Example,
		Nullable:             schema.Nullable,
		ReadOnly:             schema.ReadOnly,
		WriteOnly:            schema.WriteOnly,
		MinLength:            schema.MinLength,
		MaxLength:            schema.MaxLength,
		Pattern:              schema.Pattern,
		Minimum:              schema.Minimum,
		Maximum:              schema.Maximum,
		MinItems:             schema.MinItems,
		MaxItems:             schema.MaxItems,
		UniqueItems:          schema.UniqueItems,
	}

	parseSchemaExtensions(result)

	if schema.Ref != "" {
		result.Ref = convertRef(schema.Ref)
	}

	if schema.Items != nil {
		result.Items = convertSchema(schema.Items)
	}

	if len(schema.Properties) > 0 {
		result.Properties = make(map[string]*Schema)
		for name, prop := range schema.Properties {
			result.Properties[name] = convertSchema(prop)
		}
	}

	if len(schema.AllOf) > 0 {
		result.AllOf = make([]Schema, len(schema.AllOf))
		for i, allOf := range schema.AllOf {
			result.AllOf[i] = *convertSchema(&allOf)
		}
	}

	if len(schema.AnyOf) > 0 {
		result.AnyOf = make([]Schema, len(schema.AnyOf))
		for i, anyOf := range schema.AnyOf {
			result.AnyOf[i] = *convertSchema(&anyOf)
		}
	}

	if len(schema.OneOf) > 0 {
		result.OneOf = make([]Schema, len(schema.OneOf))
		for i, oneOf := range schema.OneOf {
			result.OneOf[i] = *convertSchema(&oneOf)
		}
	}

	if schema.Not != nil {
		result.Not = convertSchema(schema.Not)
	}

	if schema.AdditionalProperties != nil {
		result.AdditionalProperties = &AdditionalProperties{
			IsBool:    schema.AdditionalProperties.IsBool,
			BoolVal:   schema.AdditionalProperties.BoolVal,
			SchemaVal: convertSchema(schema.AdditionalProperties.SchemaVal),
		}
	}

	return result
}

func convertRef(ref string) string {
	if strings.HasPrefix(ref, "#/definitions/") {
		return strings.Replace(ref, "#/definitions/", "#/components/schemas/", 1)
	}
	if strings.HasPrefix(ref, "#/parameters/") {
		return strings.Replace(ref, "#/parameters/", "#/components/parameters/", 1)
	}
	if strings.HasPrefix(ref, "#/responses/") {
		return strings.Replace(ref, "#/responses/", "#/components/responses/", 1)
	}
	return ref
}

func convertSecurityDefinition(sec Security) Security31 {
	result := Security31{
		Type:        sec.Type,
		Description: sec.Description,
		Name:        sec.Name,
		In:          sec.In,
	}

	if sec.Type == "oauth2" {
		result.Scheme = "oauth2"
		flows := &SecurityFlows{}

		switch sec.Flow {
		case "implicit":
			flows.Implicit = &OAuthFlow{
				AuthorizationURL: sec.AuthorizationURL,
				Scopes:           sec.Scopes,
			}
		case "password":
			flows.Password = &OAuthFlow{
				TokenURL: sec.TokenURL,
				Scopes:   sec.Scopes,
			}
		case "application":
			flows.ClientCredentials = &OAuthFlow{
				TokenURL: sec.TokenURL,
				Scopes:   sec.Scopes,
			}
		case "accessCode":
			flows.AuthorizationCode = &OAuthFlow{
				AuthorizationURL: sec.AuthorizationURL,
				TokenURL:         sec.TokenURL,
				Scopes:           sec.Scopes,
			}
		}

		result.Flows = flows
	}

	if sec.Type == "apiKey" {
		result.Scheme = "apiKey"
	}

	if sec.Type == "basic" {
		result.Scheme = "basic"
	}

	return result
}
func parseSchemaExtensions(s *Schema) {
	if s == nil || s.Description == "" {
		return
	}

	// Parse @Schema(oneOf=[type1, type2, ...])
	if strings.Contains(s.Description, "@Schema(oneOf=[") {
		start := strings.Index(s.Description, "@Schema(oneOf=[") + len("@Schema(oneOf=[")
		rest := s.Description[start:]
		end := strings.Index(rest, "])")
		if end != -1 {
			oneOfStr := rest[:end]
			types := strings.Split(oneOfStr, ",")
			for _, t := range types {
				t = strings.TrimSpace(t)
				var subSchema Schema
				if t == "string" {
					subSchema = Schema{Type: "string"}
				} else if t == "object" {
					subSchema = Schema{Type: "object"}
				} else if strings.HasPrefix(t, "[]") {
					typeName := t[2:]
					subSchema = Schema{
						Type: "array",
						Items: &Schema{
							Ref: "#/components/schemas/dtos." + typeName,
						},
					}
				} else {
					subSchema = Schema{
						Ref: "#/components/schemas/dtos." + t,
					}
				}
				s.OneOf = append(s.OneOf, subSchema)
			}
			// Clear type if we have oneOf to be valid OpenAPI 3.1
			s.Type = ""
		}
	}
}
