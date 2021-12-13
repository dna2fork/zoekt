{ "contents": "{{ range $index, $ln := .Lines}}{{JsonText $ln}}\n{{end}}" }
