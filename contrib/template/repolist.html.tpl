{
   "repositories": [{{range .Repos}} {
      "name": "{{JsonText .Name}}",
      {{if .URL}}"url": "{{JsonText .URL}}",{{end}}
      "file_count": {{.Files}},
      "branches": [{{range .Branches}} "{{JsonText .Name}}",{{end}} null],
      "size": {{.Size}},
      "timestamp": "{{.IndexTime.Format "Jan 02, 2006 15:04"}}"
   }, {{end}} null],
   "document_count": {{.Stats.Documents}},
   "content_size": {{.Stats.ContentBytes}}
}
