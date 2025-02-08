package kv

type Storage interface {
	Set(key, val string)
	Get(key string) (val string, ok bool)
	GetByValue(val string) (key string, ok bool)
}
