package internal

var fileMap = map[string]bool{
	"/swagger-ui.css":                  true,
	"/swagger-ui-bundle.js":            true,
	"/swagger-ui-standalone-preset.js": true,
	"/swagger-initializer.js":          true,
	"/index.css":                       true,
	"/favicon-32x32.png":               true,
	"/favicon-16x16.png":               true,
	"/auth.swagger.json":               true,
}

func IsSwaggerFile(path string) bool {
	if _, ok := fileMap[path]; ok {
		return true
	}

	return false
}
