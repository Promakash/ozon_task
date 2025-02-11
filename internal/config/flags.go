package config

import "flag"

type AppFlags struct {
	UseRedis        bool
	UseInMemStorage bool
}

func ParseFlags() AppFlags {
	redis := flag.Bool("redis", false, "Use redis as app's cache (doesn't work if inmem is true)")
	inMem := flag.Bool("inmem", false, "Use inmemory storage instead of postgres")
	flag.Parse()

	return AppFlags{
		UseRedis:        *redis,
		UseInMemStorage: *inMem,
	}
}
