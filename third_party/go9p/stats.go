package go9p

type StatsOps interface {
	statsRegister()
	statsUnregister()
}
