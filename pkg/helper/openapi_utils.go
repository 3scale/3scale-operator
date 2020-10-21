package helper

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	// NonWordCharRegexp not word characters (== [^0-9A-Za-z_])
	NonWordCharRegexp = regexp.MustCompile(`\W`)

	// TemplateRegexp used to render openapi server URLs
	TemplateRegexp = regexp.MustCompile(`{([\w]+)}`)

	// NonAlphanumRegexp not alphanumeric
	NonAlphanumRegexp = regexp.MustCompile(`[^0-9A-Za-z]`)
)

func SystemNameFromOpenAPITitle(obj *openapi3.Swagger) string {
	openapiTitle := obj.Info.Title
	openapiTitleToLower := strings.ToLower(openapiTitle)
	return NonWordCharRegexp.ReplaceAllString(openapiTitleToLower, "_")
}

func K8sNameFromOpenAPITitle(obj *openapi3.Swagger) string {
	openapiTitle := obj.Info.Title
	openapiTitleToLower := strings.ToLower(openapiTitle)
	return NonAlphanumRegexp.ReplaceAllString(openapiTitleToLower, "")
}

func FirstServerFromOpenAPI(obj *openapi3.Swagger) *openapi3.Server {
	if obj == nil {
		return nil
	}

	// take only first server
	// From https://github.com/OAI/OpenAPI-Specification/blob/master/versions/3.0.2.md
	//   If the servers property is not provided, or is an empty array, the default value would be a Server Object with a url value of /.
	server := &openapi3.Server{
		URL:       `/`,
		Variables: map[string]*openapi3.ServerVariable{},
	}

	if len(obj.Servers) > 0 {
		server = obj.Servers[0]
	}

	return server
}

func RenderOpenAPIServerURLStr(server *openapi3.Server) (string, error) {
	if server == nil {
		return "", nil
	}

	data := &struct {
		Data map[string]string
	}{
		map[string]string{},
	}

	for variableName, variable := range server.Variables {
		data.Data[variableName] = variable.Default.(string)
	}

	urlTemplate := TemplateRegexp.ReplaceAllString(server.URL, `{{ index .Data "$1" }}`)

	tObj, err := template.New(server.URL).Parse(urlTemplate)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = tObj.Execute(&tpl, data)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func RenderOpenAPIServerURL(server *openapi3.Server) (*url.URL, error) {
	serverURLStr, err := RenderOpenAPIServerURLStr(server)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(serverURLStr)
	if err != nil {
		return nil, err
	}

	return serverURL, nil
}

type ExtendedSecurityRequirement struct {
	*openapi3.SecuritySchemeRef

	Scopes []string
}

func NewExtendedSecurityRequirement(secSchemeRef *openapi3.SecuritySchemeRef, scopes []string) *ExtendedSecurityRequirement {
	return &ExtendedSecurityRequirement{
		SecuritySchemeRef: secSchemeRef,
		Scopes:            scopes,
	}
}

func OpenAPIGlobalSecurityRequirements(openapiObj *openapi3.Swagger) []*ExtendedSecurityRequirement {
	extendedSecRequirements := make([]*ExtendedSecurityRequirement, 0)

	for _, secReq := range openapiObj.Security {
		for secReqItemName, scopes := range secReq {
			secScheme, ok := openapiObj.Components.SecuritySchemes[secReqItemName]
			if !ok {
				// should never happen. OpenAPI validation should detect this issue
				continue
			}

			extendedSecRequirements = append(extendedSecRequirements, NewExtendedSecurityRequirement(secScheme, scopes))
		}
	}

	return extendedSecRequirements
}

func MethodNameFromOpenAPIOperation(path, opVerb string, op *openapi3.Operation) string {
	sanitizedPath := NonWordCharRegexp.ReplaceAllString(path, "")

	name := fmt.Sprintf("%s%s", opVerb, sanitizedPath)

	if op.OperationID != "" {
		name = op.OperationID
	}
	return name
}

func MethodSystemNameFromOpenAPIOperation(path, opVerb string, op *openapi3.Operation) string {
	nameToLower := strings.ToLower(MethodNameFromOpenAPIOperation(path, opVerb, op))
	return NonWordCharRegexp.ReplaceAllString(nameToLower, "_")
}

func BaseURLFromOpenAPI(obj *openapi3.Swagger) (string, error) {
	server := FirstServerFromOpenAPI(obj)
	serverURL, err := RenderOpenAPIServerURL(server)
	if err != nil {
		return "", err
	}

	scheme := "https"
	if serverURL.Scheme != "" {
		scheme = serverURL.Scheme
	}
	//"#{api_spec.scheme || 'https'}://#{api_spec.host}"

	return fmt.Sprintf("%s://%s", scheme, serverURL.Host), nil
}

func BasePathFromOpenAPI(obj *openapi3.Swagger) (string, error) {
	server := FirstServerFromOpenAPI(obj)
	serverURL, err := RenderOpenAPIServerURL(server)
	if err != nil {
		return "", err
	}

	return serverURL.Path, nil
}
