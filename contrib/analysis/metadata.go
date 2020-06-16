package analysis

type Metadata struct {
   // .zoekt path; it is base folder for metadata
	BaseDir string
}

/* # folder structure
   e.g.
   /path/to/project/.git/
      /.zoekt/
         /file/
            /path/to/file/
               /@latest/...
               /commit_id/
                  /codeMap/...
                  /owner/...
                  /track/...
         /id/
         /hash/
         /commit/
            /...
         /index/
            /commit...
         /lock/
            /index.lock
*/

/* # work flow
   e.g.
   scanRepository -> file: mtime, hash, id
   scanHisotry    -> history: commit
                     (history) like (video), where (commit) like (frame)
*/
