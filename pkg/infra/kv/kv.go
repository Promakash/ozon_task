package kv

type Storage interface {
	Set(key, value string)
	Get(key string) (val string, ok bool)
}
