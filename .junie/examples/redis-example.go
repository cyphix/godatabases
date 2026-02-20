package examples

import (
	"gitea.cyphix.dev/kade/go.databases/redis"
)

func main() {
	// Initialize a new Redis database
	db, err := redis.NewRedisDatabase("localhost", "6379", "", 0, true)
	if err != nil {
		panic(err)
	}

	// Create a session store
	store, err := db.CreateRedisStore()
	if err != nil {
		panic(err)
	}

	// Use the store...
	_ = store
}
