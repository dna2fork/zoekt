package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/zoekt/analysis"
)

func (s *Server) contribTrack(p analysis.IProject, keyval url.Values, w http.ResponseWriter, r *http.Request) {
	f := keyval.Get("f")
	l := keyval.Get("l")
	cmd := keyval.Get("c")
	if f == "" {
		utilErrorStr(w, "invalid file name", 400)
		return
	}
	if cmd == "" {
		utilErrorStr(w, "unknown track command", 400)
		return
	}
	line, err := strconv.Atoi(l)
	if err != nil {
		utilErrorStr(w, "invalid line number", 400)
		return
	}
	track := analysis.NewTrackFile(p, f)
	var output string
	switch cmd {
	case "add":
		if line <= 0 {
			utilErrorStr(w, "invalid line number", 400)
			return
		}
		output, err = contribTrackAdd(track, line)
	case "del":
		if line <= 0 {
			utilErrorStr(w, "invalid line number", 400)
			return
		}
		output, err = contribTrackDel(track, line)
	case "get":
		// line > 0, return one specific track point
		// line <= 0, return all track points
		if line > 0 {
			output, err = contribTrackGetOne(track, line)
		} else {
			output, err = contribTrackGetAll(track)
		}
	case "sync":
		err = track.Sync()
		if err != nil {
			output = `{"ok":0}`
		} else {
			output = `{"ok":1}`
		}
	}
	if err != nil {
		fmt.Println("track api:", err)
		utilErrorStr(w, "internal error", 500)
		return
	}
	w.Write([]byte(output))
}

func contribTrackAdd(t *analysis.TrackFile, line int) (string, error) {
	err := t.Sync()
	if err != nil { return "", err }
	tpId, err := t.Add(line)
	if err != nil { return "", err }
	err = t.Save()
	// TODO: if fail, rollback?
	if err != nil { return "", err }
	j := fmt.Sprintf(`{"id":%d}`, tpId)
	return j, nil
}

func contribTrackDel(t *analysis.TrackFile, line int) (string, error) {
	err := t.Load()
	if err != nil { return "", err }
	err = t.Sync()
	if err != nil { return "", err }
	tp := t.GetByLine(line)
	if tp == nil { return `{"ok":1}`, nil }
	err = t.Del(tp.Id)
	if err != nil { return "", err }
	err = t.Save()
	// TODO: if fail, rollback?
	if err != nil { return "", err }
	j := fmt.Sprintf(`{"ok":1, "id":%d}`, tp.Id)
	return j, nil
}

func contribTrackGetOne(t *analysis.TrackFile, line int) (string, error) {
	err := t.Load()
	if err != nil { return "null", nil }
	tp := t.GetByLine(line)
	if tp == nil { return "null", nil }
	out := fmt.Sprintf(`{"id":%d,"line":%d}`, tp.Id, tp.Line)
	return out, nil
}

func contribTrackGetAll(t *analysis.TrackFile) (string, error) {
	tpList := make([]string, 0)
	err := t.Load()
	if err != nil { return "[]", nil }
	for _, tp := range t.Points {
		tpList = append(tpList, fmt.Sprintf(`{"id":%d,"line":%d}`, tp.Id, tp.Line))
	}
	j := fmt.Sprintf(`[%s]`, strings.Join(tpList, ","))
	return j, nil
}
