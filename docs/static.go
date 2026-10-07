package docs

import _ "embed"

//go:embed swagger.json
var ScalarSwaggerJSON []byte

//go:embed scalar.js.gz
var ScalarScriptGzip []byte
