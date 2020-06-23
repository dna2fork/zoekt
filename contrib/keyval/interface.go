package keyval

type Storage interface {
	UrlToKey(url string) string
	KeyToUrl(key string) string
	WithSubkey(key string, subkey string) string
	ListSubkey(key string) []string
	Get(key string) ([]byte, bool)
	Put(key string, value []byte) bool
	Del(key string) bool
}
