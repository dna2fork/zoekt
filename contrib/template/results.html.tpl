{
   "stat" : {
      {{if .Stats.Crashes}}"crashes": "{{JsonText .Stats.Crashes}}",{{end}}
      "query": "{{JsonText .Last.Query}}",
      "page_size": {{.Last.Num}},
      "page_size": {{len .FileMatches}},
      "ngram_match": {{.Stats.NgramMatches}},
      "doc_match": {{.Stats.MatchCount}},
      "doc_n": {{.Stats.FileCount}},
      {{if .Stats.Wait}}"wait": {{.Stats.Wait}},{{end}}
      {{if .Stats.FilesSkipped}}"file_skipped": {{.Stats.FilesSkipped}},{{end}}
      {{if .Stats.ShardsSkipped}}"shard_skipped": {{.Stats.ShardsSkipped}},{{end}}
      "file_considered": {{.Stats.FilesConsidered}},
      "file_loaded": {{.Stats.FilesLoaded}},
      "duration": "{{.Stats.Duration}}"
   },
   "hits": [
      {{range .FileMatches}}
      {
         "id": "{{JsonText .ResultID}}",
         "repository": "{{JsonText .Repo}}",
         "filename": "{{JsonText .FileName}}",
         {{if .URL}}"url": "{{JsonText .URL}}",{{end}}
      {{if .DuplicateID}}
         "duplicated": true, 
         "duplicate_id": "{{JsonText .DuplicateID}}"
      {{else}}
         {{if .Language}}"language": "{{JsonText .Language}}",{{end}}
         {{if .Branches}}"branches": [{{range .Branches}}"{{JsonText .}}",{{end}}null],{{end}}
         "matches": [{{range .Matches}} { "linenumber": {{.LineNum}}, "text": "{{range .Fragments}}{{JsonText (LimitPre 100 .Pre)}}{{JsonText .Match}}{{JsonText (LimitPost 100 .Post)}}{{end}}" },{{end}}null]
      {{end}}
      },
      {{end}}
      null
   ]
}
